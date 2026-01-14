package bigmodelasr

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"os"
	"testing"
)

// gzipDecompressData 解压缩gzip数据的辅助函数
func gzipDecompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func TestAsrWsClientWithAdvancedFeatures(t *testing.T) {
	// 创建带有完整高级功能配置的客户端
	client, err := NewAsrWsClient(
		WithAppKey(""),
		WithAccessKey(""),
		WithResourceID("volc.bigasr.sauc.duration"),
		WithEnablePunc(true), // 启用标点符号
		WithEnableItn(true),  // 启用ITN（逆文本规范化）
		WithEnableDdc(false), // 启用DDC
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// 音频设置
	audioSettings := &AudioSettings{
		Format:     "wav",
		SampleRate: 16000,
		Bits:       16,
		Channel:    1,
		Codec:      "raw",
	}

	ctx := context.Background()

	// 建立连接
	sender, err := client.Connect(ctx, audioSettings)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer sender.Close()

	// 读取音频文件
	audioData, err := os.ReadFile("output.wav")
	if err != nil {
		t.Fatalf("Failed to read audio file: %v", err)
	}

	response, err := sender.SendAudioChunk(ctx, audioData, false)
	if err != nil {
		t.Fatalf("Failed to send audio: %v", err)
	}

	t.Logf("Recognition result: %s", response.PayloadMsg.Result.Text)
}
