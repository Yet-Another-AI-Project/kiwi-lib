package resend

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/resend/resend-go/v2"
)

type ResendClient struct {
	Client *resend.Client
	opts   Options
	rdb    *redis.Client
}

func NewResendClient(options ...Option) *ResendClient {
	opts := Options{}

	for _, option := range options {
		option(&opts)
	}
	client := resend.NewClient(opts.APIKey)
	return &ResendClient{Client: client, opts: opts, rdb: opts.rdb}
}

func (c *ResendClient) SendEmail(tos []string, subject string, html string) (string, error) {
	params := &resend.SendEmailRequest{
		From:    c.opts.From,
		To:      tos,
		Html:    html,
		Subject: subject,
	}

	sent, err := c.Client.Emails.Send(params)
	if err != nil {
		return "", err
	}

	return sent.Id, nil
}

func (c *ResendClient) SendVerifyCode(to string, code string) (string, error) {
	html := fmt.Sprintf(c.opts.VerifyCodeTemplate, code)
	return c.SendEmail([]string{to}, c.opts.VerifyCodeSubject, html)
}
