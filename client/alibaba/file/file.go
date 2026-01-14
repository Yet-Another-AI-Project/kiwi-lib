package file

import (
	"context"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"io"
)

// AlibabaOssClient 是一个阿里云 OSS SDK V2 的简单包装器
type AlibabaOssClient struct {
	Client *oss.Client
}

// NewAlibabaOssClient 创建一个新的底层 OSS V2 客户端
func NewAlibabaOssClient(accessKeyID, accessKeySecret string) (*AlibabaOssClient, error) {
	// V2 SDK 的初始化流程：
	// 1. 创建凭证提供者 (Credentials Provider)
	provider := credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, "")

	// 2. 创建客户端配置 (Client Config)
	// 如果 endpoint 是 https://oss-cn-hangzhou.aliyuncs.com 这种格式，
	// 那么 region 应该是 cn-hangzhou。SDK V2 需要显式指定 Region。

	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(provider).
		WithRegion("cn-shanghai")

	// 3. 创建 OSS V2 客户端
	client := oss.NewClient(cfg)
	return &AlibabaOssClient{Client: client}, nil
}

// PutObject 封装了 OSS Go SDK V2 的 PutObject 方法
// V2 SDK 的方法现在普遍接受 context.Context 作为第一个参数
func (a *AlibabaOssClient) PutObject(ctx context.Context, bucketName, objectKey string, data io.Reader) error {
	// V2 SDK 的方法现在接收一个请求结构体作为参数
	req := &oss.PutObjectRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr(objectKey),
		Body:   data,
	}

	// 调用 V2 客户端的 PutObject 方法
	_, err := a.Client.PutObject(ctx, req)
	return err
}
