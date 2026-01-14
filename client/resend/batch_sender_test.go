package resend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"bytes"
	"os"

	"github.com/redis/go-redis/v9"
)

func TestEnqueueEmailConcurrentLatency(t *testing.T) {
	if os.Getenv("LOCAL_TEST") != "true" {
		t.Skip("skip: only run when LOCAL_TEST=true")
	}
	from := "test@futurx.cn"
	testTo := "1148562789@qq.com"
	concurrent := 100
	apiKey := "re_gwW8vqKh_7UKydcRWVAhhN6pLx4rShCSU"

	rdb := redis.NewClient(&redis.Options{
		Addr:     "139.224.251.109:6379",
		Password: "testAdmin123",
		DB:       0,
	})
	defer rdb.Close()

	rdb.Del(context.Background(), "email_batch_queue")

	client := NewResendClient(
		WithFrom(from),
		WithRedis(rdb),
		WithAPIKey(apiKey),
	)

	type sentInfo struct {
		idx      int
		sentAt   time.Time
		queuedAt time.Time
	}

	var (
		wg        sync.WaitGroup
		latencies = make([]time.Duration, concurrent)
		errors    = make([]error, concurrent)
		sentCh    = make(chan sentInfo, concurrent)
		queuedAt  = make([]time.Time, concurrent)
	)

	onError := func(err error) {
		t.Logf("send error: %v", err)
	}

	onSent := func(idx int) {
		sentCh <- sentInfo{idx: idx, sentAt: time.Now(), queuedAt: queuedAt[idx]}
	}

	// 启动批量发送协程，发送时回调onSent
	go func() {
		for range time.NewTicker(time.Second / 2).C {
			var batch []map[string]interface{}
			var idxs []int
			for i := 0; i < 100; i++ {
				data, err := rdb.LPop(context.Background(), "email_batch_queue").Result()
				if err == redis.Nil {
					break
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
				// 解析idx
				var idx int
				fmt.Sscanf(msg.Subject, "并发入队测试 #%d", &idx)
				idxs = append(idxs, idx-1)
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
				req.Header.Set("Authorization", "Bearer "+client.opts.APIKey)
				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					onError(err)
					continue
				}
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
					onError(fmt.Errorf("resend batch send failed, status: %d, err: %v", resp.StatusCode, err))
				} else {
					for _, idx := range idxs {
						onSent(idx)
					}
				}
			}
		}
	}()

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			queuedAt[idx] = time.Now()
			err := client.EnqueueEmail(
				[]string{testTo},
				fmt.Sprintf("并发入队测试 #%d", idx+1),
				fmt.Sprintf("<h1>并发入队测试 #%d</h1>", idx+1),
			)
			if err != nil {
				errors[idx] = err
			}
		}(i)
	}
	wg.Wait()

	// 收集所有发送延迟
	for i := 0; i < concurrent; i++ {
		info := <-sentCh
		latencies[info.idx] = info.sentAt.Sub(info.queuedAt)
	}

	var total time.Duration
	var min, max time.Duration
	min = time.Hour
	var errCount int
	for i, d := range latencies {
		if errors[i] != nil {
			t.Logf("goroutine %d error: %v", i, errors[i])
			errCount++
			continue
		}
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
		total += d
	}
	if concurrent-errCount > 0 {
		avg := total / time.Duration(concurrent-errCount)
		t.Logf("并发入队: %d, 成功: %d, 失败: %d, 平均发送延迟: %v, 最小: %v, 最大: %v", concurrent, concurrent-errCount, errCount, avg, min, max)
	} else {
		t.Logf("全部入队失败")
	}
}
