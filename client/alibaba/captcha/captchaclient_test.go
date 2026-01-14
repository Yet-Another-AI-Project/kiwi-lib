package captcha

import (
	"fmt"
	"testing"
)

func TestVerifyCaptcha(t *testing.T) {
	client, err := NewCaptchaClient(
		WithAccessKeySecret(""),
		WithAccessKeyId(""),
	)
	if err != nil {
		fmt.Println(err)
	}
	ok, err := client.VerifyIntelligentCaptcha("sssss", "zg8kzzfd")
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		fmt.Println("captcha校验失败")
	}
}
