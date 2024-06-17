package dtos

type SuccessData struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Success struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorData struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Error   string      `json:"error"`
	Data    interface{} `json:"data"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}
