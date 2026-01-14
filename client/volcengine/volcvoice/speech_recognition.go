package volcvoice

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/futurxlab/golanggraph/logger"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ProtocolVersion byte
type MessageType byte
type MessageTypeSpecificFlags byte
type SerializationType byte
type CompressionType byte

const (
	SuccessCode = 1000

	PROTOCOL_VERSION    = ProtocolVersion(0b0001)
	DEFAULT_HEADER_SIZE = 0b0001

	PROTOCOL_VERSION_BITS            = 4
	HEADER_BITS                      = 4
	MESSAGE_TYPE_BITS                = 4
	MESSAGE_TYPE_SPECIFIC_FLAGS_BITS = 4
	MESSAGE_SERIALIZATION_BITS       = 4
	MESSAGE_COMPRESSION_BITS         = 4
	RESERVED_BITS                    = 8

	// Message Type:
	CLIENT_FULL_REQUEST       = MessageType(0b0001)
	CLIENT_AUDIO_ONLY_REQUEST = MessageType(0b0010)
	SERVER_FULL_RESPONSE      = MessageType(0b1001)
	SERVER_ACK                = MessageType(0b1011)
	SERVER_ERROR_RESPONSE     = MessageType(0b1111)

	// Message Type Specific Flags
	NO_SEQUENCE    = MessageTypeSpecificFlags(0b0000) // no check sequence
	POS_SEQUENCE   = MessageTypeSpecificFlags(0b0001)
	NEG_SEQUENCE   = MessageTypeSpecificFlags(0b0010)
	NEG_SEQUENCE_1 = MessageTypeSpecificFlags(0b0011)

	// Message Serialization
	NO_SERIALIZATION = SerializationType(0b0000)
	JSON             = SerializationType(0b0001)
	THRIFT           = SerializationType(0b0011)
	CUSTOM_TYPE      = SerializationType(0b1111)

	// Message Compression
	NO_COMPRESSION     = CompressionType(0b0000)
	GZIP               = CompressionType(0b0001)
	CUSTOM_COMPRESSION = CompressionType(0b1111)
)

// version: b0001 (4 bits)
// header size: b0001 (4 bits)
// message type: b0001 (Full client request) (4bits)
// message type specific flags: b0000 (none) (4bits)
// message serialization method: b0001 (JSON) (4 bits)
// message compression: b0001 (gzip) (4bits)
// reserved data: 0x00 (1 byte)
var DefaultFullClientWsHeader = []byte{0x11, 0x10, 0x11, 0x00}
var DefaultAudioOnlyWsHeader = []byte{0x11, 0x20, 0x11, 0x00}
var DefaultLastAudioWsHeader = []byte{0x11, 0x22, 0x11, 0x00}

func gzipCompress(input []byte) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(input)
	w.Close()
	return b.Bytes()
}

func gzipDecompress(input []byte) []byte {
	b := bytes.NewBuffer(input)
	r, _ := gzip.NewReader(b)
	out, _ := io.ReadAll(r)
	r.Close()
	return out
}

type AsrResponse struct {
	Reqid    string   `json:"reqid"`
	Code     int      `json:"code"`
	Message  string   `json:"message"`
	Sequence int      `json:"sequence"`
	Results  []Result `json:"result,omitempty"`
}

type Result struct {
	// required
	Text       string `json:"text"`
	Confidence int    `json:"confidence"`
	// if show_language == true
	Language string `json:"language,omitempty"`
	// if show_utterances == true
	Utterances []Utterance `json:"utterances,omitempty"`
}

type Utterance struct {
	Text      string `json:"text"`
	StartTime int    `json:"start_time"`
	EndTime   int    `json:"end_time"`
	Definite  bool   `json:"definite"`
	Words     []Word `json:"words"`
	// if show_language = true
	Language string `json:"language"`
}

type Word struct {
	Text      string `json:"text"`
	StartTime int    `json:"start_time"`
	EndTime   int    `json:"end_time"`
	Pronounce string `json:"pronounce"`
	// in docs example - blank_time
	BlankDuration int `json:"blank_duration"`
}

type WsHeader struct {
	ProtocolVersion          ProtocolVersion
	DefaultHeaderSize        int
	MessageType              MessageType
	MessageTypeSpecificFlags MessageTypeSpecificFlags
	SerializationType        SerializationType
	CompressionType          CompressionType
}

type RequestAsr interface {
	requestAsr(audio_data []byte)
}

type VoiceClient struct {
	Appid   string
	Token   string
	SegSize int
	Logger  logger.ILogger
}

type AsrRequest struct {
	AudioData         []byte
	Format            string
	BoostingTableName string // 热词设置
	CorrectTableName  string // 替换词设置
	IsNluPunctuate    bool   // 是否开启标点
	IsItn             bool   // 是否开启ITN
}

func (client *VoiceClient) requestAsr(ctx context.Context, asrConfig *AsrRequest) (AsrResponse, error) {
	// set token header
	var tokenHeader = http.Header{"Authorization": []string{fmt.Sprintf("Bearer;%s", client.Token)}}
	c, _, err := websocket.DefaultDialer.Dial("wss://openspeech.bytedance.com/api/v2/asr", tokenHeader)
	if err != nil {
		return AsrResponse{}, err
	}
	defer c.Close()

	audioData := asrConfig.AudioData
	// 1. send full client request
	req := client.constructAsrRequest(asrConfig)
	payload := gzipCompress(req)
	payloadSize := len(payload)
	payloadSizeArr := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadSizeArr, uint32(payloadSize))

	fullClientMsg := make([]byte, len(DefaultFullClientWsHeader))
	copy(fullClientMsg, DefaultFullClientWsHeader)
	fullClientMsg = append(fullClientMsg, payloadSizeArr...)
	fullClientMsg = append(fullClientMsg, payload...)
	c.WriteMessage(websocket.BinaryMessage, fullClientMsg)
	_, msg, err := c.ReadMessage()
	if err != nil {
		return AsrResponse{}, err
	}
	asrResponse, err := client.parseAsrResponse(ctx, msg)
	if err != nil {
		return AsrResponse{}, err
	}

	// 3. send segment audio request
	for sentSize := 0; sentSize < len(audioData); sentSize += client.SegSize {
		lastAudio := false
		if sentSize+client.SegSize >= len(audioData) {
			lastAudio = true
		}
		// dataSlice := make([]byte, 0)
		var dataSlice []byte
		audioMsg := make([]byte, len(DefaultAudioOnlyWsHeader))
		if !lastAudio {
			dataSlice = audioData[sentSize : sentSize+client.SegSize]
			copy(audioMsg, DefaultAudioOnlyWsHeader)
		} else {
			dataSlice = audioData[sentSize:]
			copy(audioMsg, DefaultLastAudioWsHeader)
		}
		payload = gzipCompress(dataSlice)
		payloadSize := len(payload)
		payloadSizeArr := make([]byte, 4)
		binary.BigEndian.PutUint32(payloadSizeArr, uint32(payloadSize))
		audioMsg = append(audioMsg, payloadSizeArr...)
		audioMsg = append(audioMsg, payload...)
		c.WriteMessage(websocket.BinaryMessage, audioMsg)
		_, msg, err := c.ReadMessage()
		if err != nil {
			return AsrResponse{}, err
		}
		asrResponse, err = client.parseAsrResponse(ctx, msg)
		if err != nil {
			return AsrResponse{}, err
		}
	}
	return asrResponse, nil
}

func (client *VoiceClient) constructAsrRequest(asr *AsrRequest) []byte {
	reqid := uuid.New().String()
	req := make(map[string]map[string]interface{})
	req["app"] = make(map[string]interface{})
	req["app"]["appid"] = client.Appid
	req["app"]["cluster"] = "volcengine_input_common"
	req["app"]["token"] = client.Token
	req["user"] = make(map[string]interface{})
	req["user"]["uid"] = "uid"
	req["request"] = make(map[string]interface{})
	req["request"]["reqid"] = reqid
	req["request"]["nbest"] = 1
	req["request"]["workflow"] = "audio_in,resample,partition,vad,fe,decode"
	if asr.IsItn {
		req["request"]["workflow"] = fmt.Sprintf("%s,itn", req["request"]["workflow"])
	}
	if asr.IsNluPunctuate {
		req["request"]["workflow"] = fmt.Sprintf("%s,nlu_punctuate", req["request"]["workflow"])
	}
	req["request"]["result_type"] = "full"
	req["request"]["sequence"] = 1
	// 热词设置
	if asr.BoostingTableName != "" {
		req["request"]["boosting_table_name"] = asr.BoostingTableName
	}
	// 替换词设置
	if asr.CorrectTableName != "" {
		req["request"]["correct_table_name"] = asr.CorrectTableName
	}
	req["audio"] = make(map[string]interface{})
	req["audio"]["format"] = asr.Format
	req["audio"]["codec"] = "raw"
	reqStr, _ := json.Marshal(req)
	return reqStr
}

func (client *VoiceClient) parseAsrResponse(ctx context.Context, msg []byte) (AsrResponse, error) {
	client.Logger.Infof(ctx, "parseAsrResponse, msg: %x", msg)

	headerSize := msg[0] & 0x0f
	messageType := msg[1] >> 4
	serializationMethod := msg[2] >> 4
	messageCompression := msg[2] & 0x0f
	payload := msg[headerSize*4:]
	payloadMsg := make([]byte, 0)
	payloadSize := 0

	if messageType == byte(SERVER_FULL_RESPONSE) {
		payloadSize = int(int32(binary.BigEndian.Uint32(payload[0:4])))
		payloadMsg = payload[4:]
	} else if messageType == byte(SERVER_ACK) {
		seq := int32(binary.BigEndian.Uint32(payload[:4]))
		if len(payload) >= 8 {
			payloadSize = int(binary.BigEndian.Uint32(payload[4:8]))
			payloadMsg = payload[8:]
		}
		client.Logger.Infof(ctx, "SERVER_ACK, seq: %d", seq)
	} else if messageType == byte(SERVER_ERROR_RESPONSE) {
		code := int32(binary.BigEndian.Uint32(payload[:4]))
		// payloadSize = int(binary.BigEndian.Uint32(payload[4:8]))
		payloadMsg = payload[8:]
		decompressed := gzipDecompress(payloadMsg)
		client.Logger.Errorf(ctx, "SERVER_ERROR_RESPONSE, code: %d, message: %s", code, string(decompressed))
		return AsrResponse{}, errors.New(string(decompressed))
	}

	if payloadSize == 0 {
		return AsrResponse{}, errors.New("payload size is 0")
	}

	// 解压缩数据
	if messageCompression == byte(GZIP) {
		decompressed := gzipDecompress(payloadMsg)
		payloadMsg = decompressed
	}

	// 解析 JSON 数据
	var asrResponse = AsrResponse{}
	if serializationMethod == byte(JSON) {
		err := json.Unmarshal(payloadMsg, &asrResponse)
		if err != nil {
			return AsrResponse{}, err
		}
	}

	return asrResponse, nil
}
