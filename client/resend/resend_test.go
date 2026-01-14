package resend

import (
	"testing"
)

func TestSendVerifyCode(t *testing.T) {
	client := NewResendClient(
		WithAPIKey("re_Gxv4pr1Z_6crPeLfWx9jpy3PiMgQ8GZP7"),
		WithFrom("verify@mail.futurx.cn"),
		WithVerifyCodeTemplate("您的验证码是：%s"),
		WithVerifyCodeSubject("注册验证码"),
	)

	id, err := client.SendVerifyCode("lemon.yan@futurx.cn", "123456")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(id)
}
