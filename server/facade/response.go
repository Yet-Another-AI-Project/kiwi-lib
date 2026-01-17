package facade

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

type BaseResponse struct {
	Status string      `json:"status" swaggertype:"string"`
	Error  *Error      `json:"error"`
	Data   interface{} `json:"data" swaggertype:"object"`
}
