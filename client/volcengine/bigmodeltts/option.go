package bigmodeltts

import (
	"github.com/futurxlab/golanggraph/logger"
	"github.com/gorilla/websocket"
)

const (
	DefaultEndpoint   = "v3/tts/bidirection"
	DefaultResourceID = "volc.service_type.10029"
)

type Options struct {
	AppKey            string
	AccessKey         string
	Endpoint          string
	WebsocketDialer   *websocket.Dialer
	Logger            logger.ILogger
	DefaultSpeaker    string
	DefaultResourceID string
}

type Option func(*Options)

func WithEndpoint(endpoint string) Option {
	return func(o *Options) {
		o.Endpoint = endpoint
	}
}

func WithAppKey(appKey string) Option {
	return func(o *Options) {
		o.AppKey = appKey
	}
}

func WithAccessKey(accessKey string) Option {
	return func(o *Options) {
		o.AccessKey = accessKey
	}
}

func WithWebsocketDialer(dialer *websocket.Dialer) Option {
	return func(o *Options) {
		o.WebsocketDialer = dialer
	}
}

func WithLogger(logger logger.ILogger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

func WithDefaultSpeaker(speaker string) Option {
	return func(o *Options) {
		o.DefaultSpeaker = speaker
	}
}

func WithDefaultResourceID(resourceID string) Option {
	return func(o *Options) {
		o.DefaultResourceID = resourceID
	}
}
