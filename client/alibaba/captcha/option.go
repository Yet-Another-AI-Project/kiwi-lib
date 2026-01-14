package captcha

const (
	DefaultEndpoint                     = "captcha.cn-shanghai.aliyuncs.com" // 阿里云默认endpoint
	VerifyIntelligentCaptchaSuccessCode = "T001"                             // 阿里云校验码-校验成功返回码
)

type Options struct {
	AccessKeyID     string `json:"access_key_id"`     //ak
	AccessKeySecret string `json:"access_key_secret"` //sk
	EndPoint        string `json:"end_point"`         //endpoint 如captcha.cn-shanghai.aliyuncs.com
}

type Option func(options *Options)

func WithAccessKeyId(accessKeyID string) Option {
	return func(options *Options) {
		options.AccessKeyID = accessKeyID
	}
}

func WithAccessKeySecret(accessKeySecret string) Option {
	return func(options *Options) {
		options.AccessKeySecret = accessKeySecret
	}
}

func WithEndPoint(endPoint string) Option {
	return func(options *Options) {
		options.EndPoint = endPoint
	}
}
