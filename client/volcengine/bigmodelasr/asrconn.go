package bigmodelasr

import (
	"context"
	"encoding/binary"
	"errors"

	"github.com/futurxlab/golanggraph/xerror"
	"github.com/gorilla/websocket"
)

var (
	ErrWebsocketClosed = errors.New("websocket closed")
)

type AsrWsSender struct {
	conn *websocket.Conn
	seq  int32
}

func (s *AsrWsSender) SendAudioChunk(ctx context.Context, chunk []byte, last bool) (*Response, error) {
	s.seq++
	headerFlags := PosSequence
	currentSeq := s.seq

	if last {
		headerFlags = NegWithSequence
		// 最后一帧必须使用负序列号
		currentSeq = -s.seq
	}

	gzChunk, _ := gzipCompress(chunk)
	header := generateHeader(AudioOnlyRequest, byte(headerFlags), JSONSerialization, GZIPCompression)
	beforePayload := generateBeforePayload(currentSeq)
	audioRequest := append(header, beforePayload...)
	audioRequest = binary.BigEndian.AppendUint32(audioRequest, uint32(len(gzChunk)))
	audioRequest = append(audioRequest, gzChunk...)

	if err := s.conn.WriteMessage(websocket.BinaryMessage, audioRequest); err != nil {
		return nil, xerror.Wrap(err)
	}

	_, resp, err := s.conn.ReadMessage()
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			return nil, xerror.Wrap(ErrWebsocketClosed)
		}
		return nil, xerror.Wrap(err)
	}

	parsedResp, err := parseResponse(resp)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &parsedResp, nil
}

func (s *AsrWsSender) Close() error {
	return s.conn.Close()
}
