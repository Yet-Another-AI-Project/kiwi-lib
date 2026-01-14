package msgsms

import (
	"net/http"
	"time"
)

type Option func(*HuaweiCloudSMS)

type HuaweiCloudSMS struct {
	key        string
	secret     string
	sender     string
	templateId string
	signature  string
	endpoint   string
	httpClient *http.Client
}

func WithKey(key string) Option {
	return func(h *HuaweiCloudSMS) {
		h.key = key
	}
}

func WithSecret(secret string) Option {
	return func(h *HuaweiCloudSMS) {
		h.secret = secret
	}
}

func WithSender(sender string) Option {
	return func(h *HuaweiCloudSMS) {
		h.sender = sender
	}
}

func WithTemplateId(templateId string) Option {
	return func(h *HuaweiCloudSMS) {
		h.templateId = templateId
	}
}

func WithSignature(signature string) Option {
	return func(h *HuaweiCloudSMS) {
		h.signature = signature
	}
}

func WithEndpoint(endpoint string) Option {
	return func(h *HuaweiCloudSMS) {
		h.endpoint = endpoint
	}
}

func WithHttpClient(httpClient *http.Client) Option {
	return func(h *HuaweiCloudSMS) {
		h.httpClient = httpClient
	}
}

func NewHuaweiCloudSMS(opts ...Option) (*HuaweiCloudSMS, error) {

	sms := &HuaweiCloudSMS{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(sms)
	}

	return sms, nil
}
