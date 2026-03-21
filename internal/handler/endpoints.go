package handler

import (
	"fmt"
	"net/http"
	neturl "net/url"

	"github.com/Flack74/Webhook-Delivery-Platform/internal/repository"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/utils"
	"github.com/gin-gonic/gin"
)

type CreateEndpointRequest struct {
	URL         string `json:"url" binding:"required"`
	Description string `json:"description" binding:"required,min=3,max=255"`
}

type CreateEndpointResponse struct {
	ID            string `json:"id"`
	ApplicationID string `json:"application_id"`
	URL           string `json:"url"`
	Description   string `json:"description"`
	IsActive      bool   `json:"is_active"`
	Secret        string `json:"secret"`
}

type EndpointHandler struct {
	endpointRepo *repository.EndpointRepository
}

func NewEndpointHandler(endpointRepo *repository.EndpointRepository) *EndpointHandler {
	return &EndpointHandler{endpointRepo: endpointRepo}
}

func (h *EndpointHandler) CreateEndpoint(c *gin.Context) {
	applicationID, ok := authorizeApplicationPath(c)
	if !ok {
		return
	}

	var req CreateEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := validateEndpointURL(req.URL); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	secret, err := utils.GenerateSecret()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to generate endpoint secret")
		return
	}

	endpointID, err := h.endpointRepo.Insert(c.Request.Context(), applicationID, req.URL, secret, req.Description)
	if err != nil {
		respondRepositoryError(c, err)
		return
	}

	c.JSON(http.StatusCreated, CreateEndpointResponse{
		ID:            endpointID,
		ApplicationID: applicationID,
		URL:           req.URL,
		Description:   req.Description,
		IsActive:      true,
		Secret:        secret,
	})
}

func validateEndpointURL(raw string) (*neturl.URL, error) {
	parsed, err := neturl.ParseRequestURI(raw)
	if err != nil {
		return nil, err
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("url scheme must be http or https")
	}

	if parsed.Host == "" {
		return nil, fmt.Errorf("url host is required")
	}

	return parsed, nil
}
