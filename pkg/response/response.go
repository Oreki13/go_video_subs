package response

type Response struct {
	Status string `json:"status"`
	Data   any    `json:"data"`
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func OK(data any) Response {
	return Response{
		Status: "ok",
		Data:   data,
	}
}

func Fail(message string) ErrorResponse {
	return ErrorResponse{
		Status:  "error",
		Message: message,
	}
}
