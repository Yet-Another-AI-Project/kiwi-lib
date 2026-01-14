package file

type Option = func(config *Config)
type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	Endpoint        string
	Bucket          string
	Prefix          string
	CDN             string
}

func WidthAccessKeyID(accessKeyID string) Option {
	return func(config *Config) {
		config.AccessKeyID = accessKeyID
	}
}

func WithAccessKeySecret(accessKeySecret string) Option {
	return func(config *Config) {
		config.AccessKeySecret = accessKeySecret
	}
}

func WithEndpoint(endpoint string) Option {
	return func(config *Config) {
		config.Endpoint = endpoint
	}
}

func WithBucket(bucket string) Option {
	return func(config *Config) {
		config.Bucket = bucket
	}
}

func WithPrefix(prefix string) Option {
	return func(config *Config) {
		config.Prefix = prefix
	}
}

func WithCDN(cdn string) Option {
	return func(config *Config) {
		config.CDN = cdn
	}
}
