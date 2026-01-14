package msgsms

import (
	"fmt"
	"testing"
)

type sendSmsTemplateParam struct {
	Code string `json:"code"`
}

func TestSmsClient(t *testing.T) {
	client := NewSmsClient(
		WithAccessKey(""),
		WithSecretKey(""),
		WithSmsAccount("83883d4e"),
		WithSignName("向量涌现"),
		//WithTemplateID("ST_8393772b"),
		WithDefaultScene("注册验证码"),
	)

	res, err := client.SendVerifyCode("17689325441", "ST_838dbb48")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", res)

	//response, err := client.SendVerifyCode("19821136572", "注册验证码", "ST_838dbb48", 600, 6)
	//
	//if err != nil {
	//	t.Fatal(err)
	//}
	//t.Logf("%+v", response)
	//
	//success, err := client.CheckVerifyCode("19821136572", "注册验证码", "033222")
	//
	//if err != nil {
	//	t.Fatal(err)
	//}
	//t.Logf("%+v", success)

}
func TestCheckVerifyCode(t *testing.T) {
	client := NewSmsClient(
		WithAccessKey("AKLTNGZmMWJlNzE0OTkxNDI0NThhOWEzNGZjZDAyZjgzZDA"),
		WithSecretKey("WlRkbVkySTFZVGhsTXpnNE5HTTNaRGhoTVRKak1tVmpORGsyT0dJd1l6VQ=="),
		WithSmsAccount("83883d4e"),
		WithSignName("向量涌现"),
		WithDefaultScene("注册验证码"),
	)
	ok, err := client.CheckVerifyCode("17689325441", "682781")
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		fmt.Println("11校验失败")
	}
}
