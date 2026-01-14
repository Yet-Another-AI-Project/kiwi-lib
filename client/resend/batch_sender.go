package resend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"bytes"
	"net/http"

	"github.com/redis/go-redis/v9"
)

type EmailMessage struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
}

func (c *ResendClient) EnqueueEmail(tos []string, subject string, html string) error {
	msg := EmailMessage{
		From:    c.opts.From,
		To:      tos,
		Subject: subject,
		Html:    html,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.rdb.RPush(context.Background(), "email_batch_queue", data).Err()
}

func (c *ResendClient) EnqueueVerifyCode(to string, code string) error {
	html := fmt.Sprintf(c.opts.VerifyCodeTemplate, code)
	return c.EnqueueEmail([]string{to}, c.opts.VerifyCodeSubject, html)
}

func (c *ResendClient) BatchSendLoop(onError func(error)) {
	ticker := time.NewTicker(time.Second / 2)
	defer ticker.Stop()
	for range ticker.C {
		var batch []map[string]interface{}
		for i := 0; i < 100; i++ {
			data, err := c.rdb.LPop(context.Background(), "email_batch_queue").Result()
			if err == redis.Nil {
				break // 队列空
			}
			if err != nil {
				onError(err)
				break
			}
			var msg EmailMessage
			if err := json.Unmarshal([]byte(data), &msg); err != nil {
				onError(err)
				continue
			}
			batch = append(batch, map[string]interface{}{
				"from":    msg.From,
				"to":      msg.To,
				"subject": msg.Subject,
				"html":    msg.Html,
			})
		}
		if len(batch) > 0 {
			body, _ := json.Marshal(batch)
			req, err := http.NewRequest("POST", "https://api.resend.com/emails/batch", bytes.NewBuffer(body))
			if err != nil {
				onError(err)
				continue
			}
			req.Header.Set("Authorization", "Bearer "+c.opts.APIKey)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				onError(err)
				continue
			}
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
				bodyBytes, _ := io.ReadAll(resp.Body)
				onError(fmt.Errorf("resend batch send failed, status: %d, body: %s", resp.StatusCode, string(bodyBytes)))
			}
			resp.Body.Close()
		}
	}
}
