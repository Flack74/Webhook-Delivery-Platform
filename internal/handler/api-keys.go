package handler

import (
	"net/http"

	"github.com/Flack74/Webhook-Delivery-Platform/internal/repository"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/utils"
	"github.com/gin-gonic/gin"
)

type CreateAPIKeyRequest struct {
	Name string `json:"name" binding:"required,min=3,max=100"`
}

type CreateAPIKeyResponse struct {
	ID            string `json:"id"`
	ApplicationID string `json:"application_id"`
	Name          string `json:"name"`
	APIKey        string `json:"api_key"`
}

type APIKeyHandler struct {
	apiRepo *repository.APIKeyRepository
}

func NewAPIKeyHandler(apiRepo *repository.APIKeyRepository) *APIKeyHandler {
	return &APIKeyHandler{apiRepo: apiRepo}
}

func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	applicationID, ok := authorizeApplicationPath(c)
	if !ok {
		return
	}

	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	rawAPIKey, err := utils.GenerateAPIKey()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to generate api key")
		return
	}

	apiKeyID, err := h.apiRepo.Insert(c.Request.Context(), applicationID, rawAPIKey, req.Name)
	if err != nil {
		respondRepositoryError(c, err)
		return
	}

	c.JSON(http.StatusCreated, CreateAPIKeyResponse{
		ID:            apiKeyID,
		ApplicationID: applicationID,
		Name:          req.Name,
		APIKey:        rawAPIKey,
	})
}
