package obs

import (
	"io"

	"github.com/futurxlab/golanggraph/xerror"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

type HuaweiCloudObs struct {
	Client *obs.ObsClient
}

func NewHuaweiCloudObs(accessKeyID, accessKeySecret, endpoint string) (*HuaweiCloudObs, error) {

	client, err := obs.New(accessKeyID, accessKeySecret, endpoint)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &HuaweiCloudObs{Client: client}, nil
}

func (h *HuaweiCloudObs) PutObject(bucketName, objectKey string, data io.Reader) (*obs.PutObjectOutput, error) {

	input := &obs.PutObjectInput{}
	// 指定存储桶名称
	input.Bucket = bucketName
	// 指定上传对象，此处以 example/objectname 为例。
	input.Key = objectKey
	// 指定文件流
	input.Body = data

	// 流式上传本地文件
	resp, err := h.Client.PutObject(input)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return resp, nil
}

func (h *HuaweiCloudObs) GetObject(bucketName, objectKey string) (*obs.GetObjectOutput, error) {

	input := &obs.GetObjectInput{}
	input.Bucket = bucketName
	input.Key = objectKey

	resp, err := h.Client.GetObject(input)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return resp, nil
}
