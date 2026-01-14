package qrcode

// CreateQRCodeRequest 是生成小程序码接口的请求体
type CreateQRCodeRequest struct {
	Path       string  `json:"path"`        // 必须是小程序页面的 path
	EnvVersion *string `json:"env_version"` // 要打开的小程序版本。正式版为 "release"，体验版为 "trial"，开发版为 "develop"。默认是正式版。
}

type PathRequest struct {
	Type        string `json:"type"`
	SceneCodeID string `json:"scene_code_id"`
}

// APIErrorResponse 微信 API 返回的错误结构
type APIErrorResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// PartnerAPIResponse 对应三方 token 接口的返回结构
type PartnerAPIResponse struct {
	Code    string  `json:"code"`
	Msg     string  `json:"msg"`
	Data    *string `json:"data"`
	Success bool    `json:"success"`
}

// 以下是微信接口返回的一些错误码常量，用于更清晰的错误处理
const (
	ErrCodeOK                = 0
	ErrCodeAuthError         = 40001 //获取 access_token 时 AppSecret 错误，或者 access_token 无效。请开发者认真比对 AppSecret 的正确性，或查看是否正在为恰当的公众号调用接口
	ErrCodeSystemBusy        = -1    // 系统繁忙，此时请开发者稍候再试
	ErrCodeQRCodeCountLimit  = 45029 // 生成码个数总和到达最大个数限制
	ErrCodeInvalidPath       = 40159 // path 不能为空，且长度不能大于 128 字节
	ErrCodeScanCodeTimeField = 85096 // scancode_time为系统保留参数，不允许配置
	ErrInvalidArgs           = 40097 // 参数错误

)
