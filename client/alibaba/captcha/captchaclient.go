package captcha

import (
	"encoding/json"
	"errors"
	"fmt"

	captcha20230305 "github.com/alibabacloud-go/captcha-20230305/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/futurxlab/golanggraph/xerror"
)

// CaptchaClient 验证码服务接口
type CaptchaClient interface {
	VerifyIntelligentCaptcha(captchaVerifyParam, sceneID string) (bool, error)
}

// alibabaCloudClient 阿里云验证码客户端实现
type alibabaCloudClient struct {
	opts   Options
	client *captcha20230305.Client
}

// NewCaptchaClient 创建验证码客户端
func NewCaptchaClient(options ...Option) (CaptchaClient, error) {
	opts := Options{
		EndPoint: DefaultEndpoint, // 默认Endpoint
	}

	// 应用配置选项
	for _, option := range options {
		option(&opts)
	}

	// 配置阿里云客户端
	config := &openapi.Config{
		AccessKeySecret: tea.String(opts.AccessKeySecret),
		AccessKeyId:     tea.String(opts.AccessKeyID),
		Endpoint:        tea.String(opts.EndPoint),
	}

	// 初始化客户端
	client, err := captcha20230305.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("初始化客户端失败: %w", err)
	}

	return &alibabaCloudClient{
		opts:   opts,
		client: client,
	}, nil
}

// VerifyIntelligentCaptcha 验证验证码
func (c *alibabaCloudClient) VerifyIntelligentCaptcha(captchaVerifyParam, SceneID string) (bool, error) {

	request := &captcha20230305.VerifyIntelligentCaptchaRequest{
		CaptchaVerifyParam: tea.String(captchaVerifyParam),
		SceneId:            tea.String(SceneID),
	}

	// 执行验证请求
	response, err := c.client.VerifyIntelligentCaptcha(request)
	if err != nil {
		return false, handleAlibabaCloudError(err)
	}
	// response 包含服务端响应的 body 和 headers
	body, err := json.Marshal(response.Body)
	if err != nil {
		return false, xerror.Wrap(err)
	}
	fmt.Printf("verify intelligent captcha reponse body: %s\n", string(body))
	success := response.Body.Success
	verifyCode := response.Body.Result.VerifyCode
	// 检查验证结果
	return success != nil && *success && verifyCode != nil && *verifyCode == VerifyIntelligentCaptchaSuccessCode, nil
}

// handleAlibabaCloudError 处理阿里云错误
func handleAlibabaCloudError(err error) error {
	var sdkErr *tea.SDKError
	if errors.As(err, &sdkErr) {
		// 提取错误信息
		message := tea.StringValue(sdkErr.Message)
		recommend := ""

		// 解析推荐信息
		if sdkErr.Data != nil {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(tea.StringValue(sdkErr.Data)), &data); err == nil {
				if rec, ok := data["Recommend"].(string); ok {
					recommend = rec
				}
			}
		}

		// 构建详细错误信息
		errMsg := fmt.Sprintf("阿里云验证码服务错误: %s", message)
		if recommend != "" {
			errMsg += fmt.Sprintf(" | 建议: %s", recommend)
		}
		return fmt.Errorf(errMsg)
	}
	return fmt.Errorf("未知错误: %w", err)
}
