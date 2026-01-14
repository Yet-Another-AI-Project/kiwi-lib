package models

const (
	FastGPTChatMessageRoleUser      = "user"
	FastGPTChatMessageRoleAssistant = "assistant"
)

type FastGPTChatRequestMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type FastGPTChatMessageContent struct {
	Type     string    `json:"type"`
	Text     *string   `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"`
}

type FastGPTChatResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type FastGPTChatCompletionRequest struct {
	Model     string                      `json:"model"`
	Messages  []FastGPTChatRequestMessage `json:"messages"`
	Stream    bool                        `json:"stream"`
	Variables map[string]any              `json:"variables,omitempty"`
	APIKey    string                      `json:"api_key"`
	BaseURL   string                      `json:"base_url"`
}

type FastGPTChatCompletionResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Index   int                        `json:"index"`
		Message FastGPTChatResponseMessage `json:"message"`
		Delta   FastGPTChatResponseMessage `json:"delta"`
	} `json:"choices"`
}

type FastGPTChatCompletionError struct {
	Message string `json:"message"`
}
