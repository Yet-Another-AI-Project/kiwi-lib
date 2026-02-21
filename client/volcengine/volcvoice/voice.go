package volcvoice

import (
	"context"
	"fmt"
	"os"

	"github.com/Yet-Another-AI-Project/kiwi-lib/logger"
)

type VoiceService struct {
	client *VoiceClient
}

func NewVoiceService(appid string, token string, logger logger.ILogger) *VoiceService {
	return &VoiceService{
		client: &VoiceClient{
			Appid:   appid,
			Token:   token,
			SegSize: 160000,
			Logger:  logger,
		},
	}
}

func (A *VoiceService) SpeechRecognition(ctx context.Context, audioPath string, format string) (string, error) {
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return "", fmt.Errorf("read audio file error: %v", err)
	}
	asrResponse, err := A.client.requestAsr(ctx, &AsrRequest{
		AudioData: audioData,
		Format:    format,
	})
	if err != nil {
		return "", fmt.Errorf("request asr error: %v", err)
	}
	return asrResponse.Results[0].Text, nil
}

func (A *VoiceService) SpeechRecognitionByBytes(ctx context.Context, asrReq *AsrRequest) (string, error) {
	asrResponse, err := A.client.requestAsr(ctx, asrReq)
	if err != nil {
		return "", fmt.Errorf("request asr error: %v", err)
	}
	if len(asrResponse.Results) < 1 {
		return "", nil
	}
	return asrResponse.Results[0].Text, nil
}

func (A *VoiceService) StreamSynth(ctx context.Context, params *SynthParams, outFile string) (int, error) {
	if params == nil {
		return 0, fmt.Errorf("params cannot be nil")
	}
	duration, err := A.client.streamSynth(ctx, params, outFile)
	if err != nil {
		return 0, err
	}
	return duration, nil
}

func (A *VoiceService) StreamSynthByBytes(ctx context.Context, params *SynthParams) (int, []byte, error) {
	if params == nil {
		return 0, nil, fmt.Errorf("params cannot be nil")
	}
	duration, audio, err := A.client.streamSynthByBytes(ctx, params)
	if err != nil {
		return 0, nil, err
	}
	return duration, audio, nil
}

func (A *VoiceService) NonstreamSynth(ctx context.Context, params *SynthParams, outFile string) (int, error) {
	if params == nil {
		return 0, fmt.Errorf("params cannot be nil")
	}
	duration, err := A.client.nonstreamSynth(ctx, params, outFile)
	if err != nil {
		return 0, err
	}
	return duration, nil
}
