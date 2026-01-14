package msgsms

type Options struct {
	AccessKey     string `json:"access_key" yaml:"access_key"`         // AK
	SecretKey     string `json:"secret_key" yaml:"secret_key"`         // SK
	SmsAccount    string `json:"sms_account" yaml:"sms_account"`       // 短信账户
	SignName      string `json:"sign_name" yaml:"sign_name"`           // 短信签名
	TemplateID    string `json:"template_id" yaml:"template_id"`       // 模板ID
	TemplateParam string `json:"template_param" yaml:"template_param"` // 当指定的短信模板（TemplateID）存在变量时，您需要设置变量的实际值。支持传入一个或多个参数，格式示例：{"code1":"1234", "code2":"5678"}
	DefaultScene  string `json:"default_scene" yaml:"default_scene"`   // 默认使用场景
	Tag           string `json:"tag" yaml:"tag"`                       // 透传字段
}

type Option func(*Options)

func WithAccessKey(accessKey string) Option {
	return func(o *Options) {
		o.AccessKey = accessKey
	}
}

func WithSecretKey(secretKey string) Option {
	return func(o *Options) {
		o.SecretKey = secretKey
	}
}

func WithSmsAccount(smsAccount string) Option {
	return func(o *Options) {
		o.SmsAccount = smsAccount
	}
}

func WithSignName(signName string) Option {
	return func(o *Options) {
		o.SignName = signName
	}
}

func WithTemplateID(templateID string) Option {
	return func(o *Options) {
		o.TemplateID = templateID
	}
}

func WithTemplateParam(templateParam string) Option {
	return func(o *Options) {
		o.TemplateParam = templateParam
	}
}

func WithDefaultScene(defaultScene string) Option {
	return func(o *Options) {
		o.DefaultScene = defaultScene
	}
}

func WithTag(tag string) Option {
	return func(o *Options) {
		o.Tag = tag
	}
}
