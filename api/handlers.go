package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (server *Server) notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
