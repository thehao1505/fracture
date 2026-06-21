// Package middleware holds Gin middleware shared across handlers.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lukenguyen/fracture/pkg/token"
)

// Context keys under which the authenticated user's details are stored.
const (
	ContextUserID = "userID"
	ContextEmail  = "email"
)

// AuthRequired rejects requests without a valid "Authorization: Bearer <token>"
// header. On success it stashes the user id and email in the Gin context so
// downstream handlers can read them.
func AuthRequired(tokens *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed authorization header"})
			return
		}

		claims, err := tokens.Parse(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextEmail, claims.Email)
		c.Next()
	}
}
