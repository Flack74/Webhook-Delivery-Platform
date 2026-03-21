package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Flack74/Webhook-Delivery-Platform/internal/repository"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(apiKeyRepo *repository.APIKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.Fields(auth)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		apiKey, err := apiKeyRepo.GetByKey(c.Request.Context(), strings.TrimSpace(parts[1]))
		if err != nil {
			if errors.Is(err, repository.ErrAPIKeyNotFound) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate api key"})
			return
		}

		c.Set("application_id", apiKey.ApplicationID)
		c.Set("api_key_id", apiKey.ID)

		updateCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := apiKeyRepo.UpdateLastUsed(updateCtx, apiKey.ID); err != nil {
			log.Printf("update api key last_used_at failed for %s: %v", apiKey.ID, err)
		}

		c.Next()
	}
}
