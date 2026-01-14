package models

type DifyResponseMode string

const (
	DifyResponseModeStreaming DifyResponseMode = "streaming"
	DifyResponseModeBlocking  DifyResponseMode = "blocking"
)

type DifyChatMessageRequest struct {
	Inputs         struct{}              `json:"inputs"`
	ConversationID string                `json:"conversation_id"`
	Query          string                `json:"query"`
	ResponseMode   DifyResponseMode      `json:"response_mode"`
	User           string                `json:"user"`
	BaseURL        string                `json:"base_url"`
	APIKey         string                `json:"api_key"`
	Files          []DifyChatMessageFile `json:"files"`
}

type DifyChatMessageFile struct {
	Type           string `json:"type"`
	TransferMethod string `json:"transfer_method"`
	URL            string `json:"url"`
}

type DifyChatMessageResponse struct {
	Event          string `json:"event"`
	ConversationID string `json:"conversation_id"`
	Answer         string `json:"answer"`
	// Message        string `json:"message"`
}
