package bigmodelasr

import (
	"github.com/futurxlab/golanggraph/logger"
	"github.com/gorilla/websocket"
)

const (
	DefaultWSURL      = "wss://openspeech.bytedance.com/api/v3/sauc/bigmodel"
	DefaultUID        = "futurx"
	DefaultFormat     = "wav"
	DefaultRate       = 16000
	DefaultBits       = 16
	DefaultChannel    = 1
	DefaultCodec      = "raw"
	DefaultMp3SegSize = 1000
	DefaultStreaming  = true
)

type Options struct {
	WSURL             string
	UID               string
	Mp3SegSize        int
	ResourceID        string
	AppKey            string
	AccessKey         string
	Corpus            *Corpus  // 语料库配置对象
	Context           *Context // 上下文配置
	EnablePunc        bool     // 启用标点符号
	EnableItn         bool     // 启用ITN（逆文本规范化）
	EnableDdc         bool     // 启用DDC
	BoostingTableName string   // 热词表名称（向后兼容）
	CorrectTableName  string   // 替换词表名称（向后兼容）
	WebsocketDialer   *websocket.Dialer
	Logger            logger.ILogger
}

type Option func(*Options)

func WithWSURL(wsURL string) Option {
	return func(o *Options) {
		o.WSURL = wsURL
	}
}

func WithUID(uid string) Option {
	return func(o *Options) {
		o.UID = uid
	}
}

func WithMp3SegSize(mp3SegSize int) Option {
	return func(o *Options) {
		o.Mp3SegSize = mp3SegSize
	}
}

func WithResourceID(resourceID string) Option {
	return func(o *Options) {
		o.ResourceID = resourceID
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

func WithCorpusConfig(corpus *Corpus) Option {
	return func(o *Options) {
		o.Corpus = corpus
	}
}

func WithContext(context *Context) Option {
	return func(o *Options) {
		o.Context = context
	}
}

func WithEnableDdc(enableDdc bool) Option {
	return func(o *Options) {
		o.EnableDdc = enableDdc
	}
}

func WithEnablePunc(enablePunc bool) Option {
	return func(o *Options) {
		o.EnablePunc = enablePunc
	}
}

func WithEnableItn(enableItn bool) Option {
	return func(o *Options) {
		o.EnableItn = enableItn
	}
}

func WithCorpusBoostingTableName(boostingTableName string) Option {
	return func(o *Options) {
		if o.Corpus == nil {
			o.Corpus = &Corpus{}
		}
		o.Corpus.BoostingTableName = boostingTableName
	}
}
func WithCorpusBoostingTableId(correctTableId string) Option {
	return func(o *Options) {
		if o.Corpus == nil {
			o.Corpus = &Corpus{}
		}
		o.Corpus.CorrectTableId = correctTableId
	}
}

func WithCorpusCorrectTableName(correctTableName string) Option {
	return func(o *Options) {
		if o.Corpus == nil {
			o.Corpus = &Corpus{}
		}
		o.Corpus.CorrectTableName = correctTableName
	}
}
func WithCorpusCorrectTableId(correctTableId string) Option {
	return func(o *Options) {
		if o.Corpus == nil {
			o.Corpus = &Corpus{}
		}
		o.Corpus.CorrectTableId = correctTableId
	}
}
func WithCorpusContext(context string) Option {
	return func(o *Options) {
		if o.Corpus == nil {
			o.Corpus = &Corpus{}
		}
		o.Corpus.Context = context
	}
}
