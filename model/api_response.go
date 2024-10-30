package model

type APIResponseStatus string

var (
	APIResponseStatusSuccess APIResponseStatus = "success"
	APIResponseStatusError   APIResponseStatus = "error"
)

type APIResponse struct {
	StatusCode int               `json:"-"`
	Status     APIResponseStatus `json:"status"`
	Message    string            `json:"message,omitempty"`
	Data       interface{}       `json:"data,omitempty"`
}

func (a APIResponse) HTTPStatus() int {
	return a.StatusCode
}

func NewAPIResponseError(statusCode int, message string) APIResponse {
	return APIResponse{
		StatusCode: statusCode,
		Status:     APIResponseStatusError,
		Message:    message,
	}
}

func NewAPIResponseSuccess(statusCode int, data interface{}) APIResponse {
	return APIResponse{
		StatusCode: statusCode,
		Status:     APIResponseStatusSuccess,
		Data:       data,
	}
}
