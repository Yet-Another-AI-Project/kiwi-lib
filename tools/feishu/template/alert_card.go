package template

import (
	"encoding/json"
	"fmt"
)

type FeishuInteractiveRequest struct {
	MsgType string                `json:"msg_type"`
	Card    FeishuInteractiveCard `json:"card"`
}

type FeishuInteractiveCard struct {
	Header   FeishuCardHeader `json:"header"`
	Elements []interface{}    `json:"elements"`
}

type FeishuCardHeader struct {
	Title    FeishuCardTitle `json:"title"`
	Template string          `json:"template"`
}

type FeishuCardTitle struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

// BuildAlertCard 组装飞书交互式卡片消息体
func BuildAlertCard(title, fullRequestURL string, logObj map[string]interface{}) ([]byte, error) {
	logStr, _ := json.MarshalIndent(logObj, "", "  ")
	card := FeishuInteractiveCard{
		Header: FeishuCardHeader{
			Title: FeishuCardTitle{
				Tag:     "plain_text",
				Content: title,
			},
			Template: "red",
		},
		Elements: []interface{}{
			map[string]interface{}{
				"tag": "div",
				"text": map[string]interface{}{
					"tag":     "lark_md",
					"content": fmt.Sprintf("**API路径:** %s", fullRequestURL),
				},
			},
			map[string]interface{}{"tag": "hr"},
			map[string]interface{}{
				"tag": "div",
				"text": map[string]interface{}{
					"tag":     "lark_md",
					"content": "**详情:**",
				},
			},
			map[string]interface{}{
				"tag":     "markdown",
				"content": fmt.Sprintf("```json\n%s\n```", string(logStr)),
			},
		},
	}
	alert := FeishuInteractiveRequest{
		MsgType: "interactive",
		Card:    card,
	}
	return json.Marshal(alert)
}
