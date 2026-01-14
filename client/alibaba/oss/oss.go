package oss

import (
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type AliyunOss struct {
	client *oss.Client
}

func (aliyun *AliyunOss) ListObjects(bucketName, prefix string) ([]string, error) {
	bucket, err := aliyun.client.Bucket(bucketName)

	if err != nil {
		return nil, err
	}

	objects := make([]string, 0)

	continueToken := ""
	for {
		lsRes, err := bucket.ListObjectsV2(oss.Prefix(prefix), oss.ContinuationToken(continueToken))
		if err != nil {
			return nil, err
		}

		// Display the listed objects. By default, a maximum of 100 objects are returned at a time.
		for _, object := range lsRes.Objects {
			objects = append(objects, object.Key)
		}
		if lsRes.IsTruncated {
			continueToken = lsRes.NextContinuationToken
		} else {
			break
		}
	}

	return objects, nil
}

func (aliyun *AliyunOss) PutObject(bucketName, key string, data io.Reader) error {
	bucket, err := aliyun.client.Bucket(bucketName)

	if err != nil {
		return err
	}

	if err := bucket.PutObject(key, data); err != nil {
		return err
	}

	return nil
}

func (aliyun *AliyunOss) GetObject(bucketName, key string) (io.ReadCloser, error) {
	bucket, err := aliyun.client.Bucket(bucketName)

	if err != nil {
		return nil, err
	}

	body, err := bucket.GetObject(key)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func NewAliyunOss(endpoint, accessKeyID, accessKeySecret string) (*AliyunOss, error) {
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)

	if err != nil {
		return nil, err
	}

	aliyunOss := &AliyunOss{
		client: client,
	}

	return aliyunOss, nil
}
