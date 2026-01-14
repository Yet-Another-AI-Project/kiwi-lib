package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	libutils "github.com/Yet-Another-AI-Project/kiwi-lib/tools/utils"
	"github.com/futurxlab/golanggraph/xerror"
	"github.com/google/uuid"
)

type OssFileClient struct {
	alibabaOss *AlibabaOssClient
	config     *Config
}

func NewOssFileClient(opts ...Option) (*OssFileClient, error) {
	config := &Config{}
	for _, opt := range opts {
		opt(config)
	}

	if config.Endpoint == "" || config.AccessKeyID == "" || config.AccessKeySecret == "" {
		return nil, errors.New("OSS endpoint, accessKeyID, and accessKeySecret must be configured")
	}

	client, err := NewAlibabaOssClient(config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		return nil, err
	}

	return &OssFileClient{
		alibabaOss: client,
		config:     config,
	}, nil
}

// UploadFile 将文件流上传到阿里云 OSS (V2 版本)
func (c *OssFileClient) UploadFile(objectKey string, data io.Reader) (string, error) {
	fileKey := path.Join(c.config.Prefix, strings.TrimPrefix(objectKey, "/"))

	// 调用底层 V2 wrapper 的方法
	if err := c.alibabaOss.PutObject(context.Background(), c.config.Bucket, fileKey, data); err != nil {
		return "", xerror.Wrap(err)
	}

	return fmt.Sprintf("%s/%s", c.config.CDN, fileKey), nil
}

func (c *OssFileClient) UploadImageToOss(base64Image string, uri string) (fileName string, err error) {
	if base64Image == "" || !strings.HasPrefix(base64Image, "data:image/") {
		return "", errors.New("图片数据格式不正确，请提供有效的data:image/开头的base64编码图片")
	}

	data, imageType, err := libutils.Base64ImageToReader(base64Image)
	if err != nil {
		return "", err
	}

	imageFileName := fmt.Sprintf("%s.%s", uuid.NewString(), imageType)
	imagePath := path.Join(uri, imageFileName)

	url, err := c.UploadFile(imagePath, data)
	if err != nil {
		return "", err
	}
	return url, nil
}
