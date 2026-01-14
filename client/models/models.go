package models

type FuturxChatCompletionStreamResponse struct {
	FastGPT *FastGPTChatCompletionResponse
	Dify    *DifyChatMessageResponse
	Error   error
	Done    bool
}
