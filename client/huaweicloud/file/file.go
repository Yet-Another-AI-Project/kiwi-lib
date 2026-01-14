package file

import "github.com/Yet-Another-AI-Project/kiwi-lib/client/huaweicloud/obs"

type ObsFileClient struct {
	huaweiCloudObs *obs.HuaweiCloudObs
	config         *Config
}

func NewObsFileClient(opts ...Option) (*ObsFileClient, error) {
	config := &Config{}
	for _, opt := range opts {
		opt(config)
	}

	client, err := obs.NewHuaweiCloudObs(config.AccessKeyID, config.AccessKeySecret,
		config.Endpoint)

	if err != nil {
		return nil, err
	}

	return &ObsFileClient{
		huaweiCloudObs: client,
		config:         config,
	}, nil
}
