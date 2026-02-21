package feishu

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/Yet-Another-AI-Project/kiwi-lib/logger"
	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/feishu/template"
	libutils "github.com/Yet-Another-AI-Project/kiwi-lib/tools/utils"

	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/otelutils"
	"github.com/gin-gonic/gin"
)

// SendFeishuAlertAsync 发送飞书交互式卡片消息，内容参考用户给的模板
func SendFeishuAlertAsync(ctx *gin.Context, logger logger.ILogger, webhookURL, title, msg string) {
	libutils.SafeGo(ctx, logger, func() {
		scheme := "http"
		if ctx.Request.TLS != nil {
			scheme = "https"
		}
		fullRequestURL := fmt.Sprintf("%s://%s%s", scheme, ctx.Request.Host, ctx.Request.RequestURI)

		logObj := map[string]interface{}{
			"ts":        time.Now().Format(time.DateTime),
			"trace_id":  otelutils.GetTraceID(ctx),
			"msg":       msg,
			"namespace": getNamespace(ctx),
			"method":    ctx.Request.Method,
		}

		body, _ := template.BuildAlertCard(title, fullRequestURL, logObj)
		resp, err2 := http.Post(webhookURL, "application/json", bytes.NewReader(body))
		if err2 != nil {
			logger.Errorf(ctx, "SendFeishuAlertAsync Error %+v", err2)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logger.Errorf(ctx, "SendFeishuAlertAsync Error %d", resp.StatusCode)
		}
	})
}

func getNamespace(ctx *gin.Context) string {
	if v, ok := ctx.Get("namespace"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return "apiserver"
}
