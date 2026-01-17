package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	database "github.com/Sahas001/pay-on/internal/database/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	errInvalidSyncLogID   = errors.New("invalid sync log id")
	errInvalidSyncStatus  = errors.New("invalid sync status")
	errInvalidRetryCount  = errors.New("invalid retry count")
	errSyncLogNotFound    = errors.New("sync log not found")
	errInvalidTransaction = errors.New("invalid transaction id")
)

type createSyncLogRequest struct {
	TransactionID uuid.UUID `json:"transaction_id" binding:"required"`
	WalletID      uuid.UUID `json:"wallet_id" binding:"required"`
	Status        string    `json:"status" binding:"required"`
}

func (server *Server) createSyncLog(c *gin.Context) {
	var req createSyncLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	status := database.SyncStatus(req.Status)
	if !status.Valid() {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidSyncStatus))
		return
	}
	log, err := server.store.CreateSyncLog(c.Request.Context(), database.CreateSyncLogParams{
		TransactionID: req.TransactionID,
		WalletID:      req.WalletID,
		Status:        status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusCreated, log)
}

func (server *Server) listAllPendingSyncs(c *gin.Context) {
	limit, offset, ok := parseLimitOffset(c)
	if !ok {
		return
	}
	logs, err := server.store.ListAllPendingSyncs(c.Request.Context(), database.ListAllPendingSyncsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (server *Server) getSyncsNeedingRetry(c *gin.Context) {
	limit, ok := parseLimit(c, 10)
	if !ok {
		return
	}
	retries := c.Query("max_attempts")
	var attemptCount *int32
	if retries != "" {
		value, err := strconv.Atoi(retries)
		if err != nil || value < 0 {
			c.JSON(http.StatusBadRequest, errorResponse(errInvalidRetryCount))
			return
		}
		typed := int32(value)
		attemptCount = &typed
	}
	logs, err := server.store.GetSyncsNeedingRetry(c.Request.Context(), database.GetSyncsNeedingRetryParams{
		AttemptCount: attemptCount,
		Limit:        int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (server *Server) countSyncLogsByStatus(c *gin.Context) {
	status := database.SyncStatus(c.Param("status"))
	if !status.Valid() {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidSyncStatus))
		return
	}
	count, err := server.store.CountSyncLogsByStatus(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (server *Server) deleteOldSyncLogs(c *gin.Context) {
	days := c.Query("days")
	if days == "" {
		c.JSON(http.StatusBadRequest, errorResponse(errMissingQuery))
		return
	}
	if _, err := strconv.Atoi(days); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidOffset))
		return
	}
	if err := server.store.DeleteOldSyncLogs(c.Request.Context(), &days); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sync logs deleted"})
}

func (server *Server) getSyncLogByID(c *gin.Context) {
	id := c.Param("id")
	logID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidSyncLogID))
		return
	}
	log, err := server.store.GetSyncLogByID(c.Request.Context(), logID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errSyncLogNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, log)
}

type updateSyncLogStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (server *Server) updateSyncLogStatus(c *gin.Context) {
	var req updateSyncLogStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	status := database.SyncStatus(req.Status)
	if !status.Valid() {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidSyncStatus))
		return
	}
	logID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidSyncLogID))
		return
	}
	log, err := server.store.UpdateSyncLogStatus(c.Request.Context(), database.UpdateSyncLogStatusParams{
		ID:     logID,
		Status: status,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errSyncLogNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, log)
}

func (server *Server) markSettleSuccessful(c *gin.Context) {
	logID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidSyncLogID))
		return
	}
	log, err := server.store.MarkSettleSuccessful(c.Request.Context(), logID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errSyncLogNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, log)
}

type markSettleFailedRequest struct {
	ErrorMessage *string `json:"error_message"`
}

func (server *Server) markSettleFailed(c *gin.Context) {
	var req markSettleFailedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	logID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidSyncLogID))
		return
	}
	log, err := server.store.MarkSettleFailed(c.Request.Context(), database.MarkSettleFailedParams{
		ID:           logID,
		ErrorMessage: req.ErrorMessage,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errSyncLogNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, log)
}

type markSettleConflictRequest struct {
	ConflictData json.RawMessage `json:"conflict_data" binding:"required"`
}

func (server *Server) markSettleConflict(c *gin.Context) {
	var req markSettleConflictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	logID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidSyncLogID))
		return
	}
	log, err := server.store.MarkSettleConflict(c.Request.Context(), database.MarkSettleConflictParams{
		ID:           logID,
		ConflictData: req.ConflictData,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errSyncLogNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, log)
}

func (server *Server) resolveSyncConflict(c *gin.Context) {
	logID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidSyncLogID))
		return
	}
	log, err := server.store.ResolveSyncConflict(c.Request.Context(), logID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errSyncLogNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, log)
}

func (server *Server) getSyncLogsByWallet(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	limit, offset, ok := parseLimitOffset(c)
	if !ok {
		return
	}
	logs, err := server.store.GetSyncLogsByWallet(c.Request.Context(), database.GetSyncLogsByWalletParams{
		WalletID: walletID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (server *Server) listPendingSyncs(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	limit, ok := parseLimit(c, 10)
	if !ok {
		return
	}
	logs, err := server.store.ListPendingSyncs(c.Request.Context(), database.ListPendingSyncsParams{
		WalletID: walletID,
		Limit:    int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (server *Server) listFailedSyncs(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	limit, ok := parseLimit(c, 10)
	if !ok {
		return
	}
	logs, err := server.store.ListFailedSyncs(c.Request.Context(), database.ListFailedSyncsParams{
		WalletID: walletID,
		Limit:    int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (server *Server) listConflictedSyncs(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	limit, ok := parseLimit(c, 10)
	if !ok {
		return
	}
	logs, err := server.store.ListConflictedSyncs(c.Request.Context(), database.ListConflictedSyncsParams{
		WalletID: walletID,
		Limit:    int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (server *Server) getSyncStats(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	stats, err := server.store.GetSyncStats(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (server *Server) getSyncLogsByTransaction(c *gin.Context) {
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidTransaction))
		return
	}
	logs, err := server.store.GetSyncLogsByTransaction(c.Request.Context(), txID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, logs)
}
