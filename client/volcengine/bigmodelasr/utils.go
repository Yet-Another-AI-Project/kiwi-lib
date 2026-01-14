package bigmodelasr

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

type Response struct {
	IsLastPackage bool
	PayloadMsg    PayloadMsg
	ErrorMsg      string
	PayloadSize   int
	Seq           int
	Code          int
}

type PayloadMsg struct {
	AudioInfo AudioInfo `json:"audio_info"`
	Result    Result    `json:"result"`
}

type AudioInfo struct {
	Duration int `json:"duration"`
}

type Result struct {
	Additions  Additions   `json:"additions"`
	Text       string      `json:"text"`
	Utterances interface{} `json:"utterances"`
}

type Additions struct {
	LogID string `json:"log_id"`
}

func generateHeader(messageType, flags, serialMethod, compression byte) []byte {
	header := make([]byte, 4)
	header[0] = ProtocolVersion | DefaultHeaderSize
	header[1] = (messageType << 4) | flags
	header[2] = (serialMethod << 4) | compression
	header[3] = 0x00 // Reserved
	return header
}

func generateBeforePayload(seq int32) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, seq)
	return buf.Bytes()
}

func gzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return nil, err
	}
	gw.Close()
	return buf.Bytes(), nil
}

func parseResponse(data []byte) (Response, error) {
	if len(data) < 4 {
		return Response{}, errors.New("invalid response length")
	}

	res := Response{}
	headerSize := int(data[0] & 0x0F)
	messageType := data[1] >> 4
	flags := data[1] & 0x0F

	payload := data[headerSize*4:]

	if flags&0x01 != 0 {
		res.Seq = int(int32(binary.BigEndian.Uint32(payload[:4])))
		payload = payload[4:]
	}

	if flags&0x02 != 0 {
		res.IsLastPackage = true
	}

	switch messageType {
	case FullServerResponse:
		res.PayloadSize = int(binary.BigEndian.Uint32(payload[:4]))
		payload = payload[4:]
	case ServerAck:
		res.Seq = int(int32(binary.BigEndian.Uint32(payload[:4])))
		if len(payload) >= 8 {
			res.PayloadSize = int(binary.BigEndian.Uint32(payload[4:8]))
			payload = payload[8:]
		}
	case ServerErrorResponse:
		res.Code = int(binary.BigEndian.Uint32(payload[:4]))
		res.PayloadSize = int(binary.BigEndian.Uint32(payload[4:8]))
		payload = payload[8:]
	}

	if res.Code != 0 {
		res.ErrorMsg = string(payload)
	} else if data[2]>>4 == JSONSerialization {
		if data[2]&0x0F == GZIPCompression {
			gr, err := gzip.NewReader(bytes.NewReader(payload))
			if err != nil {
				return res, err
			}
			defer gr.Close()
			payload, _ = io.ReadAll(gr)
		}

		var msg PayloadMsg
		if err := json.Unmarshal(payload, &msg); err != nil {
			return res, err
		}
		res.PayloadMsg = msg
	}

	return res, nil
}
