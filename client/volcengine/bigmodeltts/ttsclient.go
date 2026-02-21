package bigmodeltts

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/gopkg/lang/fastrand"
	"github.com/gorilla/websocket"

	"github.com/Yet-Another-AI-Project/kiwi-lib/logger"
	"github.com/Yet-Another-AI-Project/kiwi-lib/xerror"
	"github.com/google/uuid"
)

type TTSWsClient struct {
	opts     Options
	protocol *BinaryProtocol
	logger   logger.ILogger
}

func NewTTSWsClient(options ...Option) (*TTSWsClient, error) {
	// TODO: default logger

	opts := Options{
		Endpoint:        DefaultEndpoint,
		WebsocketDialer: websocket.DefaultDialer,
	}

	for _, option := range options {
		option(&opts)
	}

	p := NewBinaryProtocol()
	p.SetVersion(Version1)
	p.SetHeaderSize(HeaderSize4)
	p.SetSerialization(SerializationJSON)
	p.SetCompression(CompressionNone, nil)
	p.ContainsSequence = ContainsSequence

	return &TTSWsClient{
		opts:     opts,
		protocol: p,
		logger:   opts.Logger,
	}, nil
}

type AudioResult struct {
	Data []byte
	Err  error
}

// AudioSettings 定义TTS音频输出参数
type AudioSettings struct {
	Format     string
	SampleRate int32
	Channel    int32
	SpeechRate int32
	PitchRate  int32
	Volume     int32
	Speaker    string
	ResourceID string
}

// DefaultAudioSettings 返回默认的音频设置
func DefaultAudioSettings() *AudioSettings {
	return &AudioSettings{
		Format:     "mp3",
		SampleRate: 24000,
		Channel:    1,
		SpeechRate: 0,
		PitchRate:  0,
		Volume:     100,
	}
}

// Connect establishes a new websocket connection and returns sender and receiver
func (c *TTSWsClient) Connect(ctx context.Context, audioSettings *AudioSettings) (*TTSWsSender, *TTSReceiver, error) {
	if audioSettings == nil {
		audioSettings = DefaultAudioSettings()
	}

	if audioSettings.Speaker == "" {
		audioSettings.Speaker = c.opts.DefaultSpeaker
		audioSettings.ResourceID = c.opts.DefaultResourceID
	}

	c.logger.Infof(ctx, "TTS connection, audio_settings: %+v", audioSettings)

	// Connect to websocket
	conn, err := c.dial(ctx, audioSettings.ResourceID)
	if err != nil {
		return nil, nil, xerror.Wrap(err)
	}

	// Start connection
	if err := c.startConnection(ctx, conn); err != nil {
		conn.Close()
		return nil, nil, xerror.Wrap(err)
	}

	c.logger.Infof(ctx, "TTS connection established, endpoint: %s, resource_id: %s", c.opts.Endpoint, audioSettings.ResourceID)

	// Create sender and receiver
	sender := &TTSWsSender{
		conn:          conn,
		protocol:      c.protocol,
		opts:          c.opts,
		logger:        c.logger,
		audioSettings: audioSettings,
		speaker:       audioSettings.Speaker,
	}

	receiver := NewTTSReceiver(conn, c.protocol, c.logger)

	return sender, receiver, nil
}

func (c *TTSWsClient) dial(ctx context.Context, resourceID string) (*websocket.Conn, error) {
	addr := fmt.Sprintf("wss://openspeech.bytedance.com/api/%s", c.opts.Endpoint)
	header := http.Header{
		"X-Tt-Logid":        []string{genLogID()},
		"X-Api-Resource-Id": []string{resourceID},
		"X-Api-Access-Key":  []string{c.opts.AccessKey},
		"X-Api-App-Key":     []string{c.opts.AppKey},
		"X-Api-Connect-Id":  []string{uuid.New().String()},
	}

	conn, _, err := c.opts.WebsocketDialer.DialContext(ctx, addr, header)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *TTSWsClient) startConnection(ctx context.Context, conn *websocket.Conn) error {
	msg, err := NewMessage(MsgTypeFullClient, MsgTypeFlagWithEvent)
	if err != nil {
		return xerror.Wrap(err)
	}
	msg.Event = int32(EventStartConnection)
	msg.Payload = []byte("{}")

	frame, err := c.protocol.Marshal(msg)
	if err != nil {
		return xerror.Wrap(err)
	}

	if err := conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
		return xerror.Wrap(err)
	}

	// Read ConnectionStarted message.
	mt, frame, err := conn.ReadMessage()
	if err != nil {
		return xerror.Wrap(err)
	}
	if mt != websocket.BinaryMessage && mt != websocket.TextMessage {
		return xerror.Wrap(fmt.Errorf("unexpected Websocket message type: %d", mt))
	}

	msg, _, err = Unmarshal(frame, c.protocol.ContainsSequence)
	if err != nil {
		c.logger.Infof(ctx, "StartConnection response, frame: %s", string(frame))
		return xerror.Wrap(err)
	}
	if msg.Type != MsgTypeFullServer {
		return xerror.Wrap(fmt.Errorf("unexpected ConnectionStarted message type: %s", msg.Type))
	}
	if Event(msg.Event) != EventConnectionStarted {
		return xerror.Wrap(fmt.Errorf("unexpected response event (%s) for StartConnection request", Event(msg.Event)))
	}
	c.logger.Infof(ctx, "TTS connection started, event: %s, connect_id: %s", Event(msg.Event).String(), msg.ConnectID)

	return nil
}

func genLogID() string {
	const (
		maxRandNum = 1<<24 - 1<<20
		length     = 53
		version    = "02"
		localIP    = "00000000000000000000000000000000"
	)
	ts := uint64(time.Now().UnixNano() / int64(time.Millisecond))
	r := uint64(fastrand.Uint32n(maxRandNum) + 1<<20)
	var sb strings.Builder
	sb.Grow(length)
	sb.WriteString(version)
	sb.WriteString(strconv.FormatUint(ts, 10))
	sb.WriteString(localIP)
	sb.WriteString(strconv.FormatUint(r, 16))
	return sb.String()
}
