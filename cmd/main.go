package main

import (
	"fmt"
	"iamsainadh/medassist_ai/internal/config"
	"iamsainadh/medassist_ai/internal/database"
	"iamsainadh/medassist_ai/internal/handler"
	"iamsainadh/medassist_ai/internal/ollama"
	"iamsainadh/medassist_ai/internal/repository"
	"iamsainadh/medassist_ai/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg.PORT)

	//---

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "up and running",
		})
	})

	pool, err := database.NewPostgresConnection(cfg)

	repo := repository.NewRepository(pool)

	fmt.Println(cfg.LLMModel)

	ollamaClient := ollama.NewClient(cfg.LLMUrl, cfg.EmbeddingModel, cfg.LLMModel)

	svc := service.NewService(ollamaClient, repo)

	chatSvc := service.NewChatService(repo, ollamaClient)

	handler := handler.NewHandler(svc, chatSvc)

	router.POST("/upload", handler.UploadFile)
	router.POST("/chat", handler.AskQuestion)


	router.Run(cfg.PORT)
}
