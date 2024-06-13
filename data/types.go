package data

type Message struct {
	Data string `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
