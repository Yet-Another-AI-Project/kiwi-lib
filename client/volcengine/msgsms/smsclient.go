package msgsms

import (
	"github.com/volcengine/volc-sdk-golang/service/sms"
	"strings"
)

type SmsClient interface {
	SendVerifyCode(phone, templateId string) (*sms.SmsResponse, error)
	CheckVerifyCode(phone, code string) (bool, error)
	SendSms(phones []string, templateId, templateParam string) (*sms.SmsResponse, int, error)
}

type volcanoClient struct {
	opts     Options
	instance *sms.SMS
}

func NewSmsClient(options ...Option) *volcanoClient {
	opts := Options{}

	for _, option := range options {
		option(&opts)
	}

	// 初始化火山引擎实例
	instance := sms.DefaultInstance
	instance.Client.SetAccessKey(opts.AccessKey)
	instance.Client.SetSecretKey(opts.SecretKey)

	return &volcanoClient{
		opts:     opts,
		instance: instance,
	}
}

// 发送验证码
func (c *volcanoClient) SendVerifyCode(phone, templateId string) (*sms.SmsResponse, error) {
	req := &sms.SmsVerifyCodeRequest{
		SmsAccount:  c.opts.SmsAccount,
		Sign:        c.opts.SignName,
		TemplateID:  templateId,
		PhoneNumber: phone,
		Scene:       c.opts.DefaultScene,
		ExpireTime:  600,
		CodeType:    6, // 验证码类型4/6/8位
	}

	resp, _, err := c.instance.SendVerifyCode(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// 校验验证码
func (c *volcanoClient) CheckVerifyCode(phone, code string) (bool, error) {
	req := &sms.CheckSmsVerifyCodeRequest{
		SmsAccount:  c.opts.SmsAccount,
		PhoneNumber: phone,
		Scene:       c.opts.DefaultScene,
		Code:        code,
	}

	resp, _, err := c.instance.CheckVerifyCode(req)
	if err != nil {
		return false, err
	}

	// Result 为 "0" 表示成功, "1" 错误, "2" 过期
	return resp.Result == "0", nil
}

func (c *volcanoClient) SendSms(phones []string, templateID, templateParam string) (*sms.SmsResponse, int, error) {
	req := &sms.SmsRequest{
		SmsAccount:    c.opts.SmsAccount,
		Sign:          c.opts.SignName,
		TemplateID:    templateID,
		TemplateParam: templateParam,
		PhoneNumbers:  strings.Join(phones, ","), //要求使用‘,’分隔
	}

	return c.instance.Send(req)
}
