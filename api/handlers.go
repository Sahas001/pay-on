package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	errInvalidLimit    = errors.New("invalid limit")
	errInvalidOffset   = errors.New("invalid offset")
	errMissingQuery    = errors.New("missing query parameter")
	errInvalidWalletID = errors.New("invalid wallet id")
)

func (server *Server) notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
