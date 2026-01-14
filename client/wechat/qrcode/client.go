package qrcode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/futurxlab/golanggraph/logger"
	"github.com/futurxlab/golanggraph/xerror"
)

// 定义 API 接口 URL
const (
	createQRCodeURL = "https://api.weixin.qq.com/wxa/getwxacode?access_token=%s"
)

// Client 封装了与微信小程序 API 的交互
type Client struct {
	httpClient *http.Client
	logger     logger.ILogger
}

// NewClient 是客户端的构造函数，只需要注入 logger 即可
func NewClient(logger logger.ILogger) *Client {
	if logger == nil {
		panic("WeChat client logger is nil")
	}

	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
	}
}

// GetAccessToken 直接调用三方服务 API 获取 access_token
func (c *Client) GetAccessToken(ctx context.Context, tokenEndpoint string) (string, error) {
	c.logger.Debugf(ctx, "直接从三方服务获取 access_token, endpoint: %s", tokenEndpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tokenEndpoint, nil)
	if err != nil {
		return "", fmt.Errorf("构建获取 token 请求失败: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求三方服务获取 token 失败: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.logger.Warnf(ctx, "close io error: %v", err)
		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("三方服务响应非 200: %d, body: %s", resp.StatusCode, string(body))
	}

	var partnerResp PartnerAPIResponse
	if err := json.Unmarshal(body, &partnerResp); err != nil {
		return "", fmt.Errorf("解析三方服务 token 响应失败: %w, body: %s", err, string(body))
	}

	if partnerResp.Code != "0000" || !partnerResp.Success {
		c.logger.Errorf(ctx, "三方服务返回业务错误, body: %s", string(body))
		return "", fmt.Errorf("三方服务返回错误: code=%s, msg=%s", partnerResp.Code, partnerResp.Msg)
	}

	// 如果 data 是 null，则指针为 nil
	if partnerResp.Data == nil || *partnerResp.Data == "" {
		c.logger.Errorf(ctx, "三方服务返回的 token 为空, body: %s", string(body))
		return "", fmt.Errorf("三方服务返回的 token 为空")
	}

	return *partnerResp.Data, nil
}

// createQRCodeRequest 辅助函数，用于处理微信接口特殊的响应格式
func (c *Client) createQRCodeRequest(ctx context.Context, accessToken string, payload []byte) ([]byte, error) {
	url := fmt.Sprintf(createQRCodeURL, accessToken)

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, xerror.Wrap(fmt.Errorf("请求微信API失败: %w", err))
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.logger.Warnf(ctx, "close io error: %v", err)
		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)

	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		var errResp APIErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.ErrCode != ErrCodeOK {
			return nil, xerror.Wrap(fmt.Errorf("生成小程序码错误: [%d] %s", errResp.ErrCode, errResp.ErrMsg))
		}
		return nil, xerror.Wrap(fmt.Errorf("未知的 JSON 响应: %s", string(body)))
	}

	return body, nil
}

// CreateWxaCode 调用 createwxaqrcode 接口生成小程序码
// 关键：该方法现在接收 tokenEndpoint 和 authKey 作为参数
func (c *Client) CreateWxaCode(ctx context.Context, tokenEndpoint, pathPrefix, pathType, sceneCodeID string, envVersion *string) ([]byte, error) {
	// 1. 获取 access_token (每次都调用三方服务)
	accessToken, err := c.GetAccessToken(ctx, tokenEndpoint)
	if err != nil {
		return nil, xerror.Wrap(err)
	}
	// 2. 准备请求体
	bodyBytes, marshalErr := json.Marshal(PathRequest{
		Type:        pathType,
		SceneCodeID: sceneCodeID,
	})
	if marshalErr != nil {
		return nil, xerror.Wrap(fmt.Errorf("序列化pathBody失败: %w", marshalErr))
	}
	path := fmt.Sprintf("%s%s", pathPrefix, string(bodyBytes))
	fmt.Println("path:" + path)
	reqBody := &CreateQRCodeRequest{
		Path:       path,
		EnvVersion: envVersion,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, xerror.Wrap(fmt.Errorf("序列化请求体失败: %w", err))
	}

	// 3. 调用辅助函数，发起 POST 请求
	return c.createQRCodeRequest(ctx, accessToken, payload)
}
