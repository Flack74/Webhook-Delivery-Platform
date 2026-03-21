package handler

import (
	"net/http"

	"github.com/Flack74/Webhook-Delivery-Platform/internal/repository"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/utils"
	"github.com/gin-gonic/gin"
)

type CreateApplicationRequest struct {
	Name string `json:"name" binding:"required,min=3,max=100"`
}

type CreateApplicationResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	APIKeyID  string `json:"api_key_id"`
	APIKey    string `json:"api_key"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

type ApplicationHandler struct {
	appRepo *repository.ApplicationRepository
}

func NewApplicationHandler(appRepo *repository.ApplicationRepository) *ApplicationHandler {
	return &ApplicationHandler{appRepo: appRepo}
}

func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	var req CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	rawAPIKey, err := utils.GenerateAPIKey()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to generate api key")
		return
	}

	app, apiKeyID, err := h.appRepo.CreateWithAPIKey(
		c.Request.Context(),
		req.Name,
		utils.HashAPIKey(rawAPIKey),
		"default",
	)
	if err != nil {
		respondRepositoryError(c, err)
		return
	}

	c.JSON(http.StatusCreated, CreateApplicationResponse{
		ID:        app.ID,
		Name:      app.Name,
		APIKeyID:  apiKeyID,
		APIKey:    rawAPIKey,
		IsActive:  app.IsActive,
		CreatedAt: app.CreatedAt.Format(http.TimeFormat),
	})
}
