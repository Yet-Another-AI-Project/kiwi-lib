package libutils

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image/color"
	"io"
	"strings"

	"github.com/Yet-Another-AI-Project/kiwi-lib/xerror"
	"github.com/skip2/go-qrcode"
)

func Base64ImageToReader(base64ImageStr string) (io.Reader, string, error) {

	// 检查base64图片格式
	if !strings.HasPrefix(base64ImageStr, "data:image/") {
		return nil, "", xerror.New("非法的base64图片")
	}

	// 获取图片格式
	imageType := ""
	imageTypeEnd := strings.Index(base64ImageStr, ";")
	if imageTypeEnd > 0 {
		imageType = base64ImageStr[11:imageTypeEnd] // 提取 data:image/ 和 ; 之间的格式
	}

	// 移除data:image/开头的部分
	index := strings.Index(base64ImageStr, ",")
	if index > 0 {
		base64ImageStr = base64ImageStr[index+1:]
	}

	// base64解码
	decoded, err := base64.StdEncoding.DecodeString(base64ImageStr)
	if err != nil {
		return nil, "", xerror.Wrap(err)
	}

	return bytes.NewReader(decoded), imageType, nil
}

func GenerateQRCode(content string, size int, level qrcode.RecoveryLevel) (string, error) {
	if content == "" {
		return "", errors.New("content cannot be empty")
	}

	if size <= 0 {
		size = 256 // 默认大小
	}

	// 生成二维码
	qr, err := qrcode.New(content, level)
	if err != nil {
		return "", xerror.Wrap(err)
	}

	// 设置二维码大小
	qr.DisableBorder = false
	qr.ForegroundColor = color.Black // 黑色
	qr.BackgroundColor = color.White // 白色

	// 生成PNG格式的二维码图片
	png, err := qr.PNG(size)
	if err != nil {
		return "", xerror.Wrap(err)
	}
	// 转换为base64
	base64Data := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(png))
	return base64Data, nil
}
