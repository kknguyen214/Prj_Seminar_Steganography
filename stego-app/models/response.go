package models

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ExtractedData struct {
	MessageType string `json:"message_type"`
	Content     string `json:"content"`
}
