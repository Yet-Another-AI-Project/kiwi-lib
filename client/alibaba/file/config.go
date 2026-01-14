package file

// Option 是用于配置 OssFileClient 的函数类型
type Option func(config *Config)

// Config 存储了所有连接阿里云 OSS 所需的配置
type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	Endpoint        string
	Bucket          string
	Prefix          string
	CDN             string
}

// WithAccessKeyID 设置 AccessKeyID
func WithAccessKeyID(accessKeyID string) Option {
	return func(config *Config) {
		config.AccessKeyID = accessKeyID
	}
}

// WithAccessKeySecret 设置 AccessKeySecret
func WithAccessKeySecret(accessKeySecret string) Option {
	return func(config *Config) {
		config.AccessKeySecret = accessKeySecret
	}
}

// WithEndpoint 设置 Endpoint
func WithEndpoint(endpoint string) Option {
	return func(config *Config) {
		config.Endpoint = endpoint
	}
}

// WithBucket 设置 Bucket 名称
func WithBucket(bucket string) Option {
	return func(config *Config) {
		config.Bucket = bucket
	}
}

// WithPrefix 设置上传路径的前缀
func WithPrefix(prefix string) Option {
	return func(config *Config) {
		config.Prefix = prefix
	}
}

// WithCDN 设置用于访问的 CDN 域名
func WithCDN(cdn string) Option {
	return func(config *Config) {
		config.CDN = cdn
	}
}
