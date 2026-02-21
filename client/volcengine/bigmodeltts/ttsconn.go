package bigmodeltts

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Yet-Another-AI-Project/kiwi-lib/logger"
	"github.com/Yet-Another-AI-Project/kiwi-lib/xerror"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type AudioOut struct {
	Finished bool
	Data     []byte
}

type TTSWsSender struct {
	conn          *websocket.Conn
	protocol      *BinaryProtocol
	opts          Options
	logger        logger.ILogger
	audioSettings *AudioSettings
	speaker       string
}

// StartSession starts a new TTS session
func (s *TTSWsSender) StartSession(ctx context.Context, namespace string) (string, error) {
	sessionID := uuid.New().String()
	s.logger.Debugf(context.Background(), "Starting TTS session, session_id: %s, namespace: %s", sessionID, namespace)

	req := TTSRequest{
		Event:     int32(EventStartSession),
		Namespace: namespace,
		ReqParams: &TTSReqParams{
			Speaker: s.speaker,
			AudioParams: &AudioParams{
				Format:     s.audioSettings.Format,
				SampleRate: s.audioSettings.SampleRate,
				Channel:    s.audioSettings.Channel,
				SpeechRate: s.audioSettings.SpeechRate,
				PitchRate:  s.audioSettings.PitchRate,
				Volume:     s.audioSettings.Volume,
			},
		},
	}

	payload, err := json.Marshal(&req)
	if err != nil {
		return "", xerror.Wrap(fmt.Errorf("marshal StartSession request: %w", err))
	}

	msg, err := NewMessage(MsgTypeFullClient, MsgTypeFlagWithEvent)
	if err != nil {
		return "", xerror.Wrap(err)
	}
	msg.Event = req.Event
	msg.SessionID = sessionID
	msg.Payload = payload

	frame, err := s.protocol.Marshal(msg)
	if err != nil {
		return "", xerror.Wrap(err)
	}

	if err := s.conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
		return "", xerror.Wrap(err)
	}

	s.logger.Infof(ctx, "TTS session started, session_id: %s", sessionID)
	return sessionID, nil
}

func (s *TTSWsSender) SendText(ctx context.Context, sessionID, text string) error {
	s.logger.Debugf(ctx, "Sending text to TTS, session_id: %s, text: %s", sessionID, text)

	req := TTSRequest{
		Event:     int32(EventTaskRequest),
		Namespace: "BidirectionalTTS",
		ReqParams: &TTSReqParams{
			Text:    text,
			Speaker: s.speaker,
			AudioParams: &AudioParams{
				Format:     s.audioSettings.Format,
				SampleRate: s.audioSettings.SampleRate,
				Channel:    s.audioSettings.Channel,
				SpeechRate: s.audioSettings.SpeechRate,
				PitchRate:  s.audioSettings.PitchRate,
				Volume:     s.audioSettings.Volume,
			},
		},
	}
	payload, err := json.Marshal(&req)
	if err != nil {
		return xerror.Wrap(fmt.Errorf("marshal request payload: %w", err))
	}

	msg, err := NewMessage(MsgTypeFullClient, MsgTypeFlagWithEvent)
	if err != nil {
		return xerror.Wrap(fmt.Errorf("create request message: %w", err))
	}
	msg.Event = req.Event
	msg.SessionID = sessionID
	msg.Payload = payload

	frame, err := s.protocol.Marshal(msg)
	if err != nil {
		return xerror.Wrap(fmt.Errorf("marshal request message: %w", err))
	}

	if err := s.conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
		return xerror.Wrap(fmt.Errorf("send request: %w", err))
	}

	return nil
}

func (s *TTSWsSender) FinishSession(ctx context.Context, sessionID string) error {
	msg, err := NewMessage(MsgTypeFullClient, MsgTypeFlagWithEvent)
	if err != nil {
		return err
	}
	msg.Event = int32(EventFinishSession)
	msg.SessionID = sessionID
	msg.Payload = []byte("{}")

	frame, err := s.protocol.Marshal(msg)
	if err != nil {
		return err
	}

	return s.conn.WriteMessage(websocket.BinaryMessage, frame)
}

func (s *TTSWsSender) Close() error {
	return s.conn.Close()
}

type TTSReceiver struct {
	conn     *websocket.Conn
	protocol *BinaryProtocol

	logger   logger.ILogger
	audioOut chan AudioOut
	errorOut chan error
	finished chan struct{}
}

func NewTTSReceiver(conn *websocket.Conn, protocol *BinaryProtocol, logger logger.ILogger) *TTSReceiver {
	return &TTSReceiver{
		conn:     conn,
		protocol: protocol,
		logger:   logger,
		audioOut: make(chan AudioOut, 10),
		errorOut: make(chan error, 1),
		finished: make(chan struct{}),
	}
}

func (r *TTSReceiver) Start(ctx context.Context) (<-chan AudioOut, <-chan error) {
	r.logger.Debugf(ctx, "Starting TTS receiver")
	go func() {
		defer close(r.audioOut)
		defer close(r.errorOut)

		for {
			select {
			case <-ctx.Done():
				return
			case <-r.finished:
				return
			default:
				msg, err := r.receiveMessage(ctx)
				if err != nil {
					r.errorOut <- err
					return
				}

				switch msg.Type {
				case MsgTypeFullServer:
					if Event(msg.Event) == EventSessionFinished {
						r.audioOut <- AudioOut{
							Finished: true,
						}
					}
				case MsgTypeAudioOnlyServer:
					r.audioOut <- AudioOut{
						Finished: false,
						Data:     msg.Payload,
					}
				case MsgTypeError:
					r.errorOut <- fmt.Errorf("server error: %d - %s", msg.ErrorCode, msg.Payload)
					return
				}
			}
		}
	}()

	return r.audioOut, r.errorOut
}

// WaitForEvent waits for a specific event
func (r *TTSReceiver) WaitForEvent(ctx context.Context, expectedEvent Event) error {
	msg, err := r.receiveMessage(ctx)
	if err != nil {
		return err
	}

	if Event(msg.Event) != expectedEvent {
		return fmt.Errorf("unexpected event: got %s, want %s", Event(msg.Event), expectedEvent)
	}

	return nil
}

func (r *TTSReceiver) receiveMessage(ctx context.Context) (*Message, error) {
	mt, frame, err := r.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	if mt != websocket.BinaryMessage && mt != websocket.TextMessage {
		return nil, fmt.Errorf("unexpected Websocket message type: %d", mt)
	}

	msg, _, err := Unmarshal(frame, r.protocol.ContainsSequence)
	if err != nil {
		if len(frame) > 500 {
			frame = frame[:500]
		}
		r.logger.Infof(ctx, "Data response, data: %s", string(frame))
		return nil, fmt.Errorf("unmarshal response message: %w", err)
	}
	return msg, nil
}

func (r *TTSReceiver) Close() error {
	close(r.finished)
	return r.conn.Close()
}
