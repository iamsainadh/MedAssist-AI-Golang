package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"iamsainadh/medassist_ai/internal/model"
	"io"
	"net/http"
)

type Client struct {
	BaseOllamaUrl  string
	EmbeddingModel string
	ChatModel      string
	HTTP           *http.Client
}

func NewClient(baseOllamaUrl string, embeddingModel string, chatModel string) *Client {
	return &Client{
		BaseOllamaUrl:  baseOllamaUrl,
		EmbeddingModel: embeddingModel,
		ChatModel:      chatModel,
		HTTP:           &http.Client{},
	}
}

// ------------------------------------------------------- Embed -------------------------------------------------
type embedRequest struct {
	Model string `json:"model"`
	Input any    `json:"input"`
}

type embedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func (c *Client) Embed(text string) ([]float32, error) {
	body, err := json.Marshal(embedRequest{
		Model: c.EmbeddingModel,
		Input: text,
	})

	if err != nil {
		return nil, err
	}

	res, err := c.HTTP.Post(c.BaseOllamaUrl+"/api/embed", "application/json", bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)

		return nil, fmt.Errorf("ollama embed error (%d): %s", res.StatusCode, string(b))
	}

	var er embedResponse

	if err := json.NewDecoder(res.Body).Decode(&er); err != nil {
		return nil, err
	}

	if len(er.Embeddings) == 0 {
		return nil, fmt.Errorf("ollama returned no embedding")
	}
	return er.Embeddings[0], nil
}

func (c *Client) EmbedBatch(texts []string) ([][]float32, error) {
	body, err := json.Marshal(embedRequest{
		Model: c.EmbeddingModel,
		Input: texts,
	})
	if err != nil {
		return nil, err
	}

	res, err := c.HTTP.Post(
		c.BaseOllamaUrl+"/api/embed",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("%s", b)
	}

	var resp embedResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return resp.Embeddings, nil
}

//------------------------------------------------------- Chat Related

func (c *Client) Chat(systemPrompt string, userPrompt string) (string, error) {
	reqBody := model.ChatModelRequest{
		Model: c.ChatModel,
		Messages: []model.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Stream: false,
	}

	body, err := json.Marshal(&reqBody)

	if err != nil {
		return "", err
	}

	res, err := c.HTTP.Post(c.BaseOllamaUrl+"/api/chat", "application/json", bytes.NewReader(body))

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("ollama chat error (%d): %s", res.StatusCode, string(b))
	}

	var cr model.ChatModelResponse

	if err := json.NewDecoder(res.Body).Decode(&cr); err != nil {
		return "", err
	}

	return cr.Message.Content, nil
}
