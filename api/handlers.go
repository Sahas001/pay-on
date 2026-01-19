package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	errInvalidLimit       = errors.New("invalid limit")
	errInvalidOffset      = errors.New("invalid offset")
	errMissingQuery       = errors.New("missing query parameter")
	errInvalidWalletID    = errors.New("invalid wallet id")
	errInvalidCredentials = errors.New("invalid credentials")
)

func (server *Server) notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	var pgID pgtype.UUID
	copy(pgID.Bytes[:], id[:])
	pgID.Valid = true
	return pgID
}
