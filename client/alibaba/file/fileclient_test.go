package file

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestUploadImageToOss(t *testing.T) {
	const (
		endpoint        = "https://oss-cn-shanghai.aliyuncs.com"
		accessKeyID     = "LTAI5tAneE7uHmXSFvnFV7vC"
		accessKeySecret = "jB2qzCPtcZs8qDDeByBFBQaDfxZhkq"
		bucketName      = "ciif-futurx"
		cdnHost         = "https://aliyunoss.ciif-expo.com"
		prefix          = "ciif/projectflow"
	)

	var base64ImageString = "data:image/png;base64,AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAQAQAAAAAAAAAAAAAAAAAAAAAAAB9SR//fUkf/31JH/99SR//fUkf/31JH/99SR//fUkf/31JH/99SR//fUkf/6mHbP+LXDf/fUkf/31JH/99SR//fUkf/31JH/99SR//fUkf/31JH/99SR//i104/5lwT/+RZkP/fksh/6eFaP/8/Pv/mG9N/31JH/99SR//fUkf/31JH/99SR//fUkf/31JH/+tjHL/6uHb//7+/v////////////Xx7v/8+/r//////6N/Yv99SR//fUkf/31JH/99SR//fUkf/35LIf/PvK3///////////////////////////////////////////+vj3b/fUkf/31JH/99SR//fUkf/31JH/++pI/////////////08Oz/vqSQ/8y3p///////////////////////u6CK/31JH/99SR//fUkf/31JH/+IWDL/+vn3///////s5d//iVo1/6B7XP/6+ff/8Orl/9TDtv+5nYb/nXZX/4NRKf9+SyL/fUkf/31JH/99SR//sJF3////////////onxe/35LIv+ge1z/ils1/31JH/99SR//fUkf/6qIbf/dz8T/1MK0/31JH/99SR//fUkf/8WunP///////Pv7/39MIv99SR//fUkf/31JH/99SR//fUkf/31JH//ay7///////+ri2/99SR//fUkf/31JH//Frpv///////38+/9/TCP/fUkf/31JH/99SR//fUkf/31JH/99SR//2szA///////q4dv/fUkf/31JH/99SR//r491////////////pIBi/31JH/99SR//fUkf/31JH/99SR//hVUt//j29P//////1MK1/31JH/99SR//fUkf/4dXMP/59/b//////+7o4/+MXjn/fUkf/31JH/99SR//gE0k/9XFuP///////////6aDZ/99SR//fUkf/31JH/99SR//up+I////////////9vPw/8OrmP+si3D/uZ2G/+ri2////////////97Rx/99SiD/fUkf/31JH/99SR//fUkf/31KIP/KtqX//v7+/////////////////////////////////+Xb0/+HWDH/fUkf/31JH/99SR//fUkf/31JH/99SR//fUkf/6eFaP/l2tL//v7+////////////8evn/7yhi/+BTyb/fUkf/31JH/99SR//fUkf/31JH/99SR//fUkf/31JH/99SR//fUkf/4dXMP+Uakf/jV86/31JH/99SR//fUkf/31JH/99SR//fUkf/31JH/9/SyH/f0sh/39LIf9/SyH/f0sh/39LIf9/SyH/f0sh/39LIf9/SyH/f0sh/39LIf9/SyH/f0sh/39LIf9/SyH/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="

	// 初始化客户端的方式完全一样
	client, err := NewOssFileClient(
		WithEndpoint(endpoint),
		WithAccessKeyID(accessKeyID),
		WithAccessKeySecret(accessKeySecret),
		WithBucket(bucketName),
		WithCDN(cdnHost),
		WithPrefix(prefix),
	)
	if err != nil {
		log.Fatalf("初始化 OSS 客户端失败: %v", err)
	}
	log.Println("OSS V2 客户端初始化成功！")

	// 调用业务方法的方式也完全一样
	uploadedURL, err := client.UploadImageToOss(base64ImageString, "user_avatars")
	if err != nil {
		log.Fatalf("上传图片到 OSS 失败: %v", err)
	}

	fmt.Println("图片上传成功！访问 URL:")
	fmt.Println(uploadedURL)
}

// TestUploadLocalPDFFiles 是一个用于测试上传您本地指定PDF文件的新函数
func TestUploadLocalPDFFiles(t *testing.T) {
	// --- 1. 定义OSS配置 (与您提供的测试代码一致) ---
	const (
		endpoint        = "https://oss-cn-shanghai.aliyuncs.com"
		accessKeyID     = "LTAI5tAneE7uHmXSFvnFV7vC"       // 警告: 强烈建议使用环境变量来管理密钥
		accessKeySecret = "jB2qzCPtcZs8qDDeByBFBQaDfxZhkq" // 警告: 强烈建议使用环境变量来管理密钥
		bucketName      = "ciif-futurx"
		cdnHost         = "https://aliyunoss.ciif-expo.com"
		prefix          = "ciif/projectflow/local_uploads" // 为本地上传创建一个单独的前缀
	)

	// --- 2. 指定您本地文件的绝对路径 ---
	// 这个切片包含了需要上传的所有文件的完整路径
	localFilePaths := []string{
		"/Users/kongliren/Downloads/国家会展中心分楼层导览图.pdf",
	}

	// --- 3. 初始化OSS客户端 ---
	client, err := NewOssFileClient(
		WithEndpoint(endpoint),
		WithAccessKeyID(accessKeyID),
		WithAccessKeySecret(accessKeySecret),
		WithBucket(bucketName),
		WithCDN(cdnHost),
		WithPrefix(prefix),
	)
	if err != nil {
		t.Fatalf("初始化 OSS 客户端失败: %v", err)
	}
	log.Println("OSS 客户端初始化成功！")

	// --- 4. 循环上传您指定的本地文件 ---
	for _, localPath := range localFilePaths {
		// 检查文件是否存在
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			t.Errorf("文件不存在，跳过上传: %s", localPath)
			continue
		}

		// 打开本地文件以获取 io.Reader
		file, err := os.Open(localPath)
		if err != nil {
			t.Errorf("打开本地文件 %s 失败: %v", localPath, err)
			continue // 继续尝试上传下一个文件
		}
		// 确保文件在使用后被关闭
		defer file.Close()

		// 从完整路径中提取文件名 (例如, 从 "/User/kongliren/Downloads/1.pdf" 提取出 "1.pdf")
		objectKeyInOss := filepath.Base(localPath)

		log.Printf("开始上传本地文件 '%s' 到OSS...", localPath)

		// 调用核心的 UploadFile 方法
		uploadedURL, err := client.UploadFile(objectKeyInOss, file)
		if err != nil {
			t.Errorf("上传文件 %s 到 OSS 失败: %v", localPath, err)
			continue
		}

		fmt.Printf("✅ 文件 '%s' 上传成功！\n", localPath)
		fmt.Printf("   访问 URL: %s\n", uploadedURL)
		fmt.Println("--------------------------------------------------")
	}
}
