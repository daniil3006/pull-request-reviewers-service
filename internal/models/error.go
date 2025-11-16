package models

type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}
type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetails{
			Code:    code,
			Message: message,
		},
	}
}
