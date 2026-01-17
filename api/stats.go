package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (server *Server) getSystemStats(c *gin.Context) {
	stats, err := server.store.GetSystemStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, stats)
}
