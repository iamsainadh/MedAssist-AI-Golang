package handler

import (
	"errors"
	"fmt"
	"iamsainadh/medassist_ai/internal/model"
	"iamsainadh/medassist_ai/internal/service"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	Service     *service.Service
	ChatService service.ChatService
}

func NewHandler(svc *service.Service, chatSvc service.ChatService) *Handler {
	return &Handler{
		Service:     svc,
		ChatService: chatSvc,
	}
}

func (h *Handler) UploadFile(c *gin.Context) {

	var req model.UploadRequest

	if err := c.ShouldBind(&req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			errs := make(map[string]string)

			for _, fe := range ve {
				switch fe.Tag() {
				case "required":
					errs[fe.Field()] = fe.Field() + " is required"
				default:
					errs[fe.Field()] = "Invalid value"
				}
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"errors": errs,
			})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Something Went Wrong",
		})
		return
	}

	ext := strings.ToLower(filepath.Ext(req.Document.Filename))

	if ext != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Only Pdf is Supported",
		})
		return
	}

	documentId, err := h.Service.UploadDocument(c.Request.Context(), &req)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something Went Wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"document_id": documentId,
	})
}

func (h *Handler) AskQuestion(c *gin.Context) {
	var req model.AskQuestionRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Something Went Wrong",
		})
		return
	}

	answer, err := h.ChatService.ChatAnswer(c.Request.Context(), req.DocumentID, req.Question)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something Went Wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"answer": answer,
	})
}
