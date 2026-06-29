package model

type ChatModelRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type ChatModelResponse struct {
	Message ChatMessage `json:"message"`
}

type AskQuestionRequest struct {
	Question   string `json:"question"`
	DocumentID string `json:"document_id"`
}
