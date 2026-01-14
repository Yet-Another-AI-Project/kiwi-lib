package facade

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

type BaseResponse struct {
	Status string       `json:"status" swaggertype:"string"`
	Error  *FuturxError `json:"error"`
	Data   interface{}  `json:"data" swaggertype:"object"`
}
