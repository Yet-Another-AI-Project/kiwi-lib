package volcvoice

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	enumMessageType = map[byte]string{
		11: "audio-only server response",
		12: "frontend server response",
		15: "error message from server",
	}
	enumMessageTypeSpecificFlags = map[byte]string{
		0: "no sequence number",
		1: "sequence number > 0",
		2: "last message from server (seq < 0)",
		3: "sequence number < 0",
	}
	enumMessageSerializationMethods = map[byte]string{
		0:  "no serialization",
		1:  "JSON",
		15: "custom type",
	}
	enumMessageCompression = map[byte]string{
		0:  "no compression",
		1:  "gzip",
		15: "custom compression method",
	}
)

const (
	optQuery  string = "query"
	optSubmit string = "submit"
)

type synResp struct {
	Audio  []byte
	IsLast bool
}

// version: b0001 (4 bits)
// header size: b0001 (4 bits)
// message type: b0001 (Full client request) (4bits)
// message type specific flags: b0000 (none) (4bits)
// message serialization method: b0001 (JSON) (4 bits)
// message compression: b0001 (gzip) (4bits)
// reserved data: 0x00 (1 byte)
var defaultHeader = []byte{0x11, 0x10, 0x11, 0x00}

// SynthParams 语音合成参数
type SynthParams struct {
	Text        string  // 要合成的文本
	VoiceType   string  // 音色类型
	Cluster     string  // 集群
	Format      string  // 音频格式
	VolumeRatio float64 // 音量比例，默认1.0
	SpeedRatio  float64 // 语速比例，默认1.0
	Operation   string  // 操作类型：query(一次性合成) 或 submit(流式合成)
}

func (client *VoiceClient) setupInput(params *SynthParams) []byte {
	reqID := uuid.New().String()
	reqParams := make(map[string]map[string]interface{})
	reqParams["app"] = make(map[string]interface{})
	reqParams["app"]["appid"] = client.Appid
	reqParams["app"]["token"] = client.Token
	reqParams["app"]["cluster"] = params.Cluster
	reqParams["user"] = make(map[string]interface{})
	reqParams["user"]["uid"] = "uid"
	reqParams["audio"] = make(map[string]interface{})
	reqParams["audio"]["voice_type"] = params.VoiceType
	reqParams["audio"]["encoding"] = params.Format

	// 设置默认值
	speedRatio := 1.0
	if params.SpeedRatio != 0 {
		speedRatio = params.SpeedRatio
	}
	volumeRatio := 1.0
	if params.VolumeRatio != 0 {
		volumeRatio = params.VolumeRatio
	}

	reqParams["audio"]["speed_ratio"] = speedRatio
	reqParams["audio"]["volume_ratio"] = volumeRatio
	reqParams["request"] = make(map[string]interface{})
	reqParams["request"]["reqid"] = reqID
	reqParams["request"]["text"] = params.Text
	reqParams["request"]["text_type"] = "plain"
	reqParams["request"]["operation"] = params.Operation
	resStr, _ := json.Marshal(reqParams)
	return resStr
}

func (client *VoiceClient) parseTTSResponse(ctx context.Context, res []byte) (resp synResp, err error) {
	// protoVersion := res[0] >> 4
	headSize := res[0] & 0x0f
	messageType := res[1] >> 4
	messageTypeSpecificFlags := res[1] & 0x0f
	// serializationMethod := res[2] >> 4
	messageCompression := res[2] & 0x0f
	// reserve := res[3]
	headerExtensions := res[4 : headSize*4]
	payload := res[headSize*4:]

	// client.Logger.Debug(ctx, "parseTTSResponse", zap.String("protoVersion", fmt.Sprintf("%x", protoVersion)), zap.Int("protoVersion", int(protoVersion)))
	// client.Logger.Debug(ctx, "parseTTSResponse", zap.String("headSize", fmt.Sprintf("%x", headSize)), zap.Int("headSize", int(headSize)*4))
	// client.Logger.Debug(ctx, "parseTTSResponse", zap.String("messageType", fmt.Sprintf("%x", messageType)), zap.String("messageType", enumMessageType[messageType]))
	// client.Logger.Debug(ctx, "parseTTSResponse", zap.String("messageTypeSpecificFlags", fmt.Sprintf("%x", messageTypeSpecificFlags)), zap.String("messageTypeSpecificFlags", enumMessageTypeSpecificFlags[messageTypeSpecificFlags]))
	// client.Logger.Debug(ctx, "parseTTSResponse", zap.String("serializationMethod", fmt.Sprintf("%x", serializationMethod)), zap.String("serializationMethod", enumMessageSerializationMethods[serializationMethod]))
	// client.Logger.Debug(ctx, "parseTTSResponse", zap.String("messageCompression", fmt.Sprintf("%x", messageCompression)), zap.String("messageCompression", enumMessageCompression[messageCompression]))
	// client.Logger.Debug(ctx, "parseTTSResponse", zap.Int("reserve", int(reserve)))
	if headSize != 1 {
		client.Logger.Debugf(ctx, "parseTTSResponse, headerExtensions: %x", headerExtensions)
	}
	// audio-only server response
	if messageType == 0xb {
		// no sequence number as ACK
		if messageTypeSpecificFlags == 0 {
			client.Logger.Debugf(ctx, "Payload size, size: 0")
		} else {
			sequenceNumber := int32(binary.BigEndian.Uint32(payload[0:4]))
			// payloadSize := int32(binary.BigEndian.Uint32(payload[4:8]))
			payload = payload[8:]
			resp.Audio = append(resp.Audio, payload...)
			// client.Logger.Debug(ctx, "parseTTSResponse", zap.Int32("sequenceNumber", sequenceNumber))
			// client.Logger.Debug(ctx, "parseTTSResponse", zap.Int32("payloadSize", payloadSize))
			if sequenceNumber < 0 {
				resp.IsLast = true
			}
		}
	} else if messageType == 0xf {
		code := int32(binary.BigEndian.Uint32(payload[0:4]))
		errMsg := payload[8:]
		if messageCompression == 1 {
			errMsg = gzipDecompress(errMsg)
		}
		client.Logger.Debugf(ctx, "parseTTSResponse, code: %d", code)
		client.Logger.Debugf(ctx, "parseTTSResponse, errMsg: %s", string(errMsg))
		err = errors.New(string(errMsg))
		return
	} else if messageType == 0xc {
		// msgSize = int32(binary.BigEndian.Uint32(payload[0:4]))
		payload = payload[4:]
		if messageCompression == 1 {
			payload = gzipDecompress(payload)
		}
		client.Logger.Debugf(ctx, "parseTTSResponse, payload: %s", string(payload))
	} else {
		client.Logger.Debugf(ctx, "parseTTSResponse, messageType: %x", messageType)
		err = errors.New("wrong message type")
		return
	}
	return
}

// 一次性合成
func (client *VoiceClient) nonstreamSynth(ctx context.Context, params *SynthParams, outFile string) (int, error) {
	if params.Operation == "" {
		params.Operation = optQuery
	}
	input := client.setupInput(params)
	client.Logger.Infof(ctx, "nonstreamSynth, input: %s", string(input))
	input = gzipCompress(input)
	payloadSize := len(input)
	payloadArr := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadArr, uint32(payloadSize))

	clientRequest := make([]byte, len(defaultHeader))
	copy(clientRequest, defaultHeader)
	clientRequest = append(clientRequest, payloadArr...)
	clientRequest = append(clientRequest, input...)

	u := url.URL{Scheme: "wss", Host: "openspeech.bytedance.com", Path: "/api/v1/tts/ws_binary"}
	header := http.Header{"Authorization": []string{fmt.Sprintf("Bearer;%s", client.Token)}}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return 0, fmt.Errorf("dial err: %v", err)
	}
	defer c.Close()
	err = c.WriteMessage(websocket.BinaryMessage, clientRequest)
	if err != nil {
		return 0, fmt.Errorf("write message fail, err:: %v", err)
	}
	_, message, err := c.ReadMessage()
	if err != nil {
		return 0, fmt.Errorf("read message fail, err:: %v", err)
	}
	resp, err := client.parseTTSResponse(ctx, message)
	if err != nil {
		return 0, fmt.Errorf("parse response fail, err: %v", err)
	}

	duration := client.calculateAudioDuration(resp.Audio)

	err = os.WriteFile(outFile, resp.Audio, 0644)
	if err != nil {
		return 0, fmt.Errorf("write audio to fail fail, err: %v", err)
	}
	return int(duration + 1), nil
}

// 流式合成
func (client *VoiceClient) streamSynth(ctx context.Context, params *SynthParams, outFile string) (int, error) {
	if params.Operation == "" {
		params.Operation = optSubmit
	}
	input := client.setupInput(params)
	client.Logger.Infof(ctx, "streamSynth, input: %s", string(input))
	input = gzipCompress(input)
	payloadSize := len(input)
	payloadArr := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadArr, uint32(payloadSize))

	clientRequest := make([]byte, len(defaultHeader))
	copy(clientRequest, defaultHeader)
	clientRequest = append(clientRequest, payloadArr...)
	clientRequest = append(clientRequest, input...)

	u := url.URL{Scheme: "wss", Host: "openspeech.bytedance.com", Path: "/api/v1/tts/ws_binary"}
	header := http.Header{"Authorization": []string{fmt.Sprintf("Bearer;%s", client.Token)}}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return 0, fmt.Errorf("dial err: %v", err)
	}
	defer c.Close()
	if err := c.WriteMessage(websocket.BinaryMessage, clientRequest); err != nil {
		return 0, fmt.Errorf("write message fail, err:: %v", err)
	}
	var audio []byte
	for {
		var message []byte
		_, message, err = c.ReadMessage()
		if err != nil {
			return 0, fmt.Errorf("read message fail, err: %v", err)
		}
		resp, err := client.parseTTSResponse(ctx, message)
		if err != nil {
			return 0, fmt.Errorf("parse response fail, err: %v", err)
		}
		audio = append(audio, resp.Audio...)
		if resp.IsLast {
			break
		}
	}

	duration := client.calculateAudioDuration(audio)

	err = os.WriteFile(outFile, audio, 0644)
	if err != nil {
		return 0, fmt.Errorf("write audio to fail fail, err: %v", err)
	}
	return int(duration + 1), nil
}

func (client *VoiceClient) streamSynthByBytes(ctx context.Context, params *SynthParams) (int, []byte, error) {
	if params.Operation == "" {
		params.Operation = optSubmit
	}
	input := client.setupInput(params)
	client.Logger.Infof(ctx, "streamSynthByBytes, input: %s", string(input))
	input = gzipCompress(input)
	payloadSize := len(input)
	payloadArr := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadArr, uint32(payloadSize))

	clientRequest := make([]byte, len(defaultHeader))
	copy(clientRequest, defaultHeader)
	clientRequest = append(clientRequest, payloadArr...)
	clientRequest = append(clientRequest, input...)

	u := url.URL{Scheme: "wss", Host: "openspeech.bytedance.com", Path: "/api/v1/tts/ws_binary"}
	header := http.Header{"Authorization": []string{fmt.Sprintf("Bearer;%s", client.Token)}}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return 0, nil, fmt.Errorf("dial err: %v", err)
	}
	defer c.Close()
	if err := c.WriteMessage(websocket.BinaryMessage, clientRequest); err != nil {
		return 0, nil, fmt.Errorf("write message fail, err:: %v", err)
	}
	var audio []byte
	for {
		var message []byte
		_, message, err = c.ReadMessage()
		if err != nil {
			return 0, nil, fmt.Errorf("read message fail, err: %v", err)
		}
		resp, err := client.parseTTSResponse(ctx, message)
		if err != nil {
			return 0, nil, fmt.Errorf("parse response fail, err: %v", err)
		}
		audio = append(audio, resp.Audio...)
		if resp.IsLast {
			break
		}
	}

	duration := client.calculateAudioDuration(audio)

	return int(duration + 1), audio, nil
}

func (client *VoiceClient) calculateAudioDuration(pcmData []byte) float64 {
	sampleRate := 24000
	bitDepth := 16
	channels := 1
	duration := float64(len(pcmData)) / (float64(sampleRate) * float64(bitDepth) / 8 * float64(channels))
	return duration
}
