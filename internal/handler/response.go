package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// respondInternal logs the real error server-side and returns a generic message
// to the client, so internal details (DB errors, driver messages, etc.) never
// leak out over the API.
func respondInternal(c *gin.Context, err error) {
	log.Printf("internal error on %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
