package resend

import (
	"github.com/redis/go-redis/v9"
)

type Options struct {
	APIKey             string `json:"api_key" yaml:"api_key"`
	From               string `json:"from" yaml:"from"`
	VerifyCodeTemplate string `json:"verify_code_template" yaml:"verify_code_template"`
	VerifyCodeSubject  string `json:"verify_code_subject" yaml:"verify_code_subject"`
	rdb                *redis.Client
}

type Option func(*Options)

func WithAPIKey(apiKey string) Option {
	return func(o *Options) {
		o.APIKey = apiKey
	}
}

func WithFrom(from string) Option {
	return func(o *Options) {
		o.From = from
	}
}

func WithVerifyCodeTemplate(verifyCodeTemplate string) Option {
	return func(o *Options) {
		o.VerifyCodeTemplate = verifyCodeTemplate
	}
}

func WithVerifyCodeSubject(verifyCodeSubject string) Option {
	return func(o *Options) {
		o.VerifyCodeSubject = verifyCodeSubject
	}
}

// WithRedis 注入redis client
func WithRedis(rdb *redis.Client) Option {
	return func(o *Options) {
		o.rdb = rdb
	}
}
