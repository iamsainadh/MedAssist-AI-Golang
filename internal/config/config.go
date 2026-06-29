package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PORT           string
	LLMUrl         string
	DBUri          string
	EmbeddingModel string
	LLMModel       string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()

	if err != nil {
		return &Config{}, err
	}

	return &Config{
		PORT:           ":" + os.Getenv("PORT"),
		DBUri:          os.Getenv("DBUri"),
		LLMUrl:         os.Getenv("LLM_URL"),
		EmbeddingModel: os.Getenv("EMBEDDING_MODEL"),
		LLMModel:       os.Getenv("LLM_MODEL"),
	}, nil
}
