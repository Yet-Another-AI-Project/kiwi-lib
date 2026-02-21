package file

import (
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	libutils "github.com/Yet-Another-AI-Project/kiwi-lib/tools/utils"
	"github.com/Yet-Another-AI-Project/kiwi-lib/xerror"
	"github.com/google/uuid"
)

func (c *ObsFileClient) UploadImageToObs(base64Image string, uri string) (fileName string, err error) {
	if base64Image != "" && strings.HasPrefix(base64Image, "data:image/") {
		data, imageType, err := libutils.Base64ImageToReader(base64Image)
		if err != nil {
			return "", err
		}

		imageUri := fmt.Sprintf("/%s.%s", uuid.New(), imageType)
		imageUri = path.Join("/", uri, imageUri)

		url, err := c.UploadFile(imageUri, data)
		if err != nil {
			return "", err
		}
		return url, nil
	}
	return "", errors.New("图片数据格式不正确，请提供有效的data:image/开头的base64编码图片")
}
func (c *ObsFileClient) UploadFile(objectKey string, data io.Reader) (string, error) {
	fileKey := path.Join(c.config.Prefix, "/", objectKey)

	if _, err := c.huaweiCloudObs.PutObject(c.config.Bucket, fileKey, data); err != nil {
		return "", xerror.Wrap(err)
	}

	return fmt.Sprintf("%s%s", c.config.CDN, path.Join("/", fileKey)), nil
}
