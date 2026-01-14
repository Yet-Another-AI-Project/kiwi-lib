package libutils

import (
	"encoding/base64"
	"github.com/skip2/go-qrcode"
	"os"
	"strings"
	"testing"
)

// =================================================================
// 目的 :作为“生成器”使用，生成并保存二维码文件
// 可以在这里修改内容、大小等参数，然后运行这个测试来生成需要的二维码。
// =================================================================

func TestHelper_GenerateAndSaveQRCode(t *testing.T) {
	// --- 在这里修改你要生成的内容 ---
	//shortContentToGenerate := "http://hiwaic-projectflow-api.futurx.cc/link/01977d50"
	shortContentToGenerate := "http://localhost:9090/link/01977d50"
	//originContentToGenerate := "https://waic-test.futurx.cn/#/chat/01973ad9-235e-74d9-9090-71415e87da2e?type=opening&scene_code_id=01977d50-6b51-75f5-93b6-4eb42578bb07"
	outputFilename := "/Users/kongliren/Downloads/generated_qrcode_short_link_low_256_local.png" // 加下划线前缀，以便区分
	// ---------------------------------

	t.Logf("开始生成二维码，内容: '%s'", shortContentToGenerate)

	// 1. 调用函数生成 Base64 编码的二维码
	base64Data, err := GenerateQRCode(shortContentToGenerate, 256, qrcode.Low)
	if err != nil {
		// 如果生成失败，立即终止测试并报告错误
		t.Fatalf("生成二维码失败: %v", err)
	}

	t.Logf("二维码 Base64 数据已生成。")
	// t.Logf("Base64 String: %s", base64Data) // 如果需要可以取消本行注释来打印完整的 Base64 字符串

	// 2. 将 Base64 数据解码为原始的图片字节
	// 首先移除 "data:image/png;base64," 这个前缀
	parts := strings.Split(base64Data, ",")
	if len(parts) != 2 {
		t.Fatalf("Base64 字符串格式不正确")
	}
	// 解码
	imageBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("解码 Base64 数据失败: %v", err)
	}

	// 3. 将图片字节写入文件
	err = os.WriteFile(outputFilename, imageBytes, 0644)
	if err != nil {
		t.Fatalf("保存二维码图片到文件失败: %v", err)
	}

	// 4. 打印成功信息
	// 使用 t.Logf 打印的信息只有在 `go test -v` 模式下才会显示
	t.Logf("✅ 二维码已成功保存到文件: %s", outputFilename)
	t.Logf("请在项目目录下查找并打开该文件。")
}
