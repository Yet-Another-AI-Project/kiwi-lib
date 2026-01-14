package file

import (
	"fmt"
	"os"
	"path"
	"testing"
)

// TestUploadFile 是一个测试函数，用于上传一个本地的文件。
func TestUploadMP4File(t *testing.T) {
	const (
		accessKeyID     = "MJ5FDJVQWIET3U8LATIJ"
		accessKeySecret = "jUsHBpNbr869L5x1fjY5JtrqUpYZKiOSiGUyXism"
		endpoint        = "https://obs.cn-east-3.myhuaweicloud.com"
		bucket          = "futurx"
		cdnDomain       = "https://futurx.obs.cn-east-3.myhuaweicloud.com"
		uploadPrefix    = "test/projectflow"
	)
	client, err := NewObsFileClient(
		WidthAccessKeyID(accessKeyID),
		WithAccessKeySecret(accessKeySecret),
		WithEndpoint(endpoint),
		WithBucket(bucket),
		WithCDN(cdnDomain),
		WithPrefix(uploadPrefix),
	)
	if err != nil {
		t.Fatalf("创建 ObsFileClient 失败: %v", err)
	}

	const localFilePath = "/Users/kongliren/Downloads/new首页720.gif"
	file, err := os.Open(localFilePath)
	if err != nil {
		t.Fatalf("打开本地文件 '%s' 失败: %v", localFilePath, err)
	}
	defer file.Close()

	const objectKey = "test-Think.gif"

	t.Logf("开始上传文件 '%s' 到 OBS...", localFilePath)

	returnedURL, err := client.UploadFile(objectKey, file)
	if err != nil {
		t.Errorf("UploadFile 方法执行失败: %v", err)
		return
	}

	// 构造期望的 URL 格式进行比对。
	expectedURL := fmt.Sprintf("%s%s", cdnDomain, path.Join("/", uploadPrefix, objectKey))

	if returnedURL == "" {
		t.Error("返回的 URL 为空，期望得到一个有效的 URL。")
	} else if returnedURL != expectedURL {
		t.Errorf("返回的 URL 与期望不符。\n期望: %s\n实际: %s", expectedURL, returnedURL)
	} else {
		t.Logf("文件上传成功！")
		t.Logf("返回的 URL 是: %s", returnedURL)
	}
}
