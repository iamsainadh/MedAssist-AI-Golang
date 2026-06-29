package service

import (
	"context"
	"fmt"
	"iamsainadh/medassist_ai/internal/ollama"
	"iamsainadh/medassist_ai/internal/repository"
)

type ChatService interface {
	ChatAnswer(ctx context.Context, documentId string, question string) (string, error)
}
type chatService struct {
	repo   repository.Repository
	client *ollama.Client
}

func NewChatService(repo repository.Repository, client *ollama.Client) ChatService {
	return &chatService{
		repo:   repo,
		client: client,
	}
}

const systemPrompt = `
You are an AI Medical Report Assistant.
Rules:
1. Answer ONLY using the provided medical report context.
2. If the answer is not present in the report, reply:
"I couldn't find this information in the uploaded medical report."
3. Do NOT answer questions unrelated to the uploaded medical report.
4. Do NOT answer questions about politics, programming, history, sports, movies, mathematics, or any non-medical topic.
5. Do NOT make assumptions or hallucinate.
6. If the user asks for a diagnosis, clearly state that you are not a doctor and provide only an explanation based on the report.
7. Keep answers concise and easy to understand.
`

func (c *chatService) ChatAnswer(ctx context.Context, documentId string, question string) (string, error) {

	embedQuestion, err := c.client.Embed(question)

	if err != nil {
		return "", err
	}

	results, err := c.repo.SearchChunk(ctx, documentId, embedQuestion, 5)

	if err != nil {
		return "", fmt.Errorf("search chunks: %w", err)
	}
	if len(results) == 0 {
		return "", fmt.Errorf("no data found for this report")
	}

	userPrompt := fmt.Sprintf(`
		Medical Report Context:
		%s
		User Question:
		%s
		Answer only from the above report.
	`, results, question)

	response, err := c.client.Chat(systemPrompt, userPrompt)

	if err != nil {
		return "", err
	}

	return response, nil
}
