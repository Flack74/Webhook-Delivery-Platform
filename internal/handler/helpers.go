package handler

import (
	"errors"
	"net/http"

	"github.com/Flack74/Webhook-Delivery-Platform/internal/repository"
	"github.com/gin-gonic/gin"
)

const (
	contextApplicationIDKey = "application_id"
	contextAPIKeyIDKey      = "api_key_id"
)

func respondError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}

func respondRepositoryError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrApplicationNotFound),
		errors.Is(err, repository.ErrAPIKeyNotFound),
		errors.Is(err, repository.ErrEndpointNotFound),
		errors.Is(err, repository.ErrEventNotFound),
		errors.Is(err, repository.ErrDeliveryNotFound):
		respondError(c, http.StatusNotFound, err.Error())
	default:
		respondError(c, http.StatusInternalServerError, "internal server error")
	}
}

func authenticatedApplicationID(c *gin.Context) (string, bool) {
	value, exists := c.Get(contextApplicationIDKey)
	if !exists {
		respondError(c, http.StatusInternalServerError, "missing authenticated application")
		return "", false
	}

	applicationID, ok := value.(string)
	if !ok || applicationID == "" {
		respondError(c, http.StatusInternalServerError, "invalid authenticated application")
		return "", false
	}

	return applicationID, true
}

func authorizeApplicationPath(c *gin.Context) (string, bool) {
	applicationID, ok := authenticatedApplicationID(c)
	if !ok {
		return "", false
	}

	if pathID := c.Param("id"); pathID != "" && pathID != applicationID {
		respondError(c, http.StatusForbidden, "application scope mismatch")
		return "", false
	}

	return applicationID, true
}
