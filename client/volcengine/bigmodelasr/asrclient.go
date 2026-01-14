package bigmodelasr

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"

	"github.com/futurxlab/golanggraph/xerror"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	ProtocolVersion     = 0x01 << 4
	DefaultHeaderSize   = 0x01
	FullClientRequest   = 0x01
	AudioOnlyRequest    = 0x02
	FullServerResponse  = 0x09
	ServerAck           = 0x0B
	ServerErrorResponse = 0x0F
	NoSequence          = 0x00
	PosSequence         = 0x01
	NegWithSequence     = 0x03
	JSONSerialization   = 0x01
	GZIPCompression     = 0x01
)

type AsrWsClient struct {
	opts Options
}

func NewAsrWsClient(options ...Option) (*AsrWsClient, error) {

	opts := Options{
		WSURL:           DefaultWSURL,
		UID:             DefaultUID,
		Mp3SegSize:      DefaultMp3SegSize,
		EnablePunc:      true,  // 默认启用标点符号
		EnableItn:       false, // 默认不启用ITN
		WebsocketDialer: websocket.DefaultDialer,
	}

	for _, option := range options {
		option(&opts)
	}

	return &AsrWsClient{
		opts: opts,
	}, nil
}

type RequestParams struct {
	User    *User          `json:"user"`
	Audio   *AudioSettings `json:"audio"`
	Request *Request       `json:"request"`
}

type User struct {
	UID string `json:"uid"`
}

type AudioSettings struct {
	Format     string `json:"format"`
	SampleRate int    `json:"sample_rate"`
	Bits       int    `json:"bits"`
	Channel    int    `json:"channel"`
	Codec      string `json:"codec"`
}

type Request struct {
	ModelName  string  `json:"model_name"`
	EnablePunc bool    `json:"enable_punc"`
	EnableItn  bool    `json:"enable_itn,omitempty"` // 启用ITN（逆文本规范化）
	EnableDdc  bool    `json:"enable_ddc,omitempty"` // 启用DDC
	Corpus     *Corpus `json:"corpus,omitempty"`     // 语料库配置对象
}

// Corpus 语料库配置
type Corpus struct {
	BoostingTableID   string `json:"boosting_table_id,omitempty"`   // 热词表ID
	BoostingTableName string `json:"boosting_table_name,omitempty"` // 热词表名称
	CorrectTableName  string `json:"correct_table_name,omitempty"`  // 替换词表名称
	CorrectTableId    string `json:"correct_table_id,omitempty"`
	Context           string `json:"context,omitempty"` // 上下文
}

// Context 上下文配置
type Context struct {
	ContextType string        `json:"context_type,omitempty"` // 上下文类型，如 "dialog_ctx"
	ContextData []ContextData `json:"context_data,omitempty"` // 上下文数据
}

// ContextData 上下文数据项
type ContextData struct {
	Text string `json:"text"` // 上下文文本
}

func (c *AsrWsClient) constructFullRequest(audio *AudioSettings) []byte {
	req := RequestParams{
		User:  &User{UID: c.opts.UID},
		Audio: audio,
		Request: &Request{
			ModelName:  "bigmodel",
			EnablePunc: c.opts.EnablePunc, // 使用配置中的标点设置
			EnableItn:  c.opts.EnableItn,  // 使用配置中的ITN设置
			EnableDdc:  c.opts.EnableDdc,  // 使用配置中的DDC设置
			Corpus:     c.opts.Corpus,     // 使用配置中的语料库设置
		},
	}

	jsonData, _ := json.Marshal(req)
	gzData, _ := gzipCompress(jsonData)
	return gzData
}

func (c *AsrWsClient) Connect(ctx context.Context, audioSettings *AudioSettings) (*AsrWsSender, error) {
	headers := http.Header{}
	connectID := uuid.New().String()
	headers.Add("X-Api-Connect-Id", connectID)
	headers.Add("X-Api-Resource-Id", c.opts.ResourceID)
	headers.Add("X-Api-App-Key", c.opts.AppKey)
	headers.Add("X-Api-Access-Key", c.opts.AccessKey)

	conn, resp, err := c.opts.WebsocketDialer.DialContext(ctx, c.opts.WSURL, headers)
	if err != nil {
		// 添加详细的错误信息
		var statusCode int
		var responseBody string
		if resp != nil {
			statusCode = resp.StatusCode
			if resp.Body != nil {
				bodyBytes, readErr := io.ReadAll(resp.Body)
				if readErr == nil {
					responseBody = string(bodyBytes)
				}
			}
		}

		c.opts.Logger.Errorf(ctx, "WebSocket connection failed: status_code: %d, response_body: %s, %v", statusCode, responseBody, err)
		return nil, xerror.Wrap(err)
	}

	// Send initial request
	seq := int32(1)
	payload := c.constructFullRequest(audioSettings)
	header := generateHeader(FullClientRequest, PosSequence, JSONSerialization, GZIPCompression)
	beforePayload := generateBeforePayload(seq)
	fullRequest := append(header, beforePayload...)
	fullRequest = binary.BigEndian.AppendUint32(fullRequest, uint32(len(payload)))
	fullRequest = append(fullRequest, payload...)

	if err := conn.WriteMessage(websocket.BinaryMessage, fullRequest); err != nil {
		return nil, xerror.Wrap(err)
	}

	// Read initial response
	_, respData, err := conn.ReadMessage()
	if err != nil {
		return nil, xerror.Wrap(err)
	}
	parsedResp, err := parseResponse(respData)
	if err != nil {
		return nil, err
	}

	c.opts.Logger.Infof(ctx, "connected to ASR server, asr_log_id: %s", parsedResp.PayloadMsg.Result.Additions.LogID)

	asrWsSender := &AsrWsSender{
		conn: conn,
		seq:  1,
	}

	return asrWsSender, nil
}
