package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	database "github.com/Sahas001/pay-on/internal/database/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	errInvalidTransactionID   = errors.New("invalid transaction id")
	errInvalidStatus          = errors.New("invalid transaction status")
	errInvalidType            = errors.New("invalid transaction type")
	errInvalidConnectionType  = errors.New("invalid connection type")
	errInvalidNonce           = errors.New("invalid nonce")
	errInvalidAmount          = errors.New("invalid amount")
	errTransactionNotFound    = errors.New("transaction not found")
	errInvalidMetadataPayload = errors.New("invalid metadata payload")
)

type createTransactionRequest struct {
	FromWalletID   uuid.UUID       `json:"from_wallet_id" binding:"required"`
	ToWalletID     uuid.UUID       `json:"to_wallet_id" binding:"required"`
	Amount         pgtype.Numeric  `json:"amount" binding:"required"`
	Currency       string          `json:"currency"`
	Type           string          `json:"type"`
	Status         string          `json:"status"`
	Signature      string          `json:"signature" binding:"required"`
	Nonce          int64           `json:"nonce" binding:"required"`
	ConnectionType string          `json:"connection_type"`
	Description    *string         `json:"description"`
	Metadata       json.RawMessage `json:"metadata"`
	TransactionAt  *time.Time      `json:"transaction_at"`
}

func (server *Server) createTransaction(c *gin.Context) {
	var req createTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	txType := database.TransactionType(req.Type)
	if req.Type == "" {
		txType = database.TransactionTypeP2p
	}
	if !txType.Valid() {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidType))
		return
	}

	txStatus := database.TransactionStatus(req.Status)
	if req.Status == "" {
		txStatus = database.TransactionStatusPending
	}
	if !txStatus.Valid() {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidStatus))
		return
	}

	var connType database.NullConnectionType
	if req.ConnectionType != "" {
		typed := database.ConnectionType(req.ConnectionType)
		if !typed.Valid() {
			c.JSON(http.StatusBadRequest, errorResponse(errInvalidConnectionType))
			return
		}
		connType = database.NullConnectionType{ConnectionType: typed, Valid: true}
	}

	metadata := req.Metadata
	if len(metadata) == 0 {
		metadata = []byte(`{}`)
	}

	txTime := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	if req.TransactionAt != nil {
		txTime = pgtype.Timestamptz{Time: req.TransactionAt.UTC(), Valid: true}
	}

	transaction, err := server.store.CreateTransaction(c.Request.Context(), database.CreateTransactionParams{
		FromWalletID:   req.FromWalletID,
		ToWalletID:     req.ToWalletID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Type:           txType,
		Status:         txStatus,
		Signature:      req.Signature,
		Nonce:          req.Nonce,
		ConnectionType: connType,
		Description:    req.Description,
		Metadata:       metadata,
		TransactionAt:  txTime,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusCreated, transaction)
}

func (server *Server) searchTransactions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, errorResponse(errMissingQuery))
		return
	}

	limit, offset, ok := parseLimitOffset(c)
	if !ok {
		return
	}

	results, err := server.store.SearchTransactions(c.Request.Context(), database.SearchTransactionsParams{
		Column1: &query,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, results)
}

func (server *Server) getRecentTransactions(c *gin.Context) {
	limit, ok := parseLimit(c, 10)
	if !ok {
		return
	}
	results, err := server.store.GetRecentTransactions(c.Request.Context(), int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, results)
}

func (server *Server) listTransactionsByStatus(c *gin.Context) {
	status := database.TransactionStatus(c.Param("status"))
	if !status.Valid() {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidStatus))
		return
	}

	limit, offset, ok := parseLimitOffset(c)
	if !ok {
		return
	}

	results, err := server.store.ListTransactionsByStatus(c.Request.Context(), database.ListTransactionsByStatusParams{
		Status: status,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, results)
}

func (server *Server) listUnsyncedTransactions(c *gin.Context) {
	limit, offset, ok := parseLimitOffset(c)
	if !ok {
		return
	}
	results, err := server.store.ListUnsyncedTransactions(c.Request.Context(), database.ListUnsyncedTransactionsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, results)
}

func (server *Server) countPendingTransactions(c *gin.Context) {
	count, err := server.store.CountPendingTransactions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (server *Server) getTransactionsByConnectionType(c *gin.Context) {
	connType := database.ConnectionType(c.Param("type"))
	if !connType.Valid() {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidConnectionType))
		return
	}
	after := c.Query("after")
	if after == "" {
		c.JSON(http.StatusBadRequest, errorResponse(errMissingQuery))
		return
	}
	t, err := time.Parse(time.RFC3339, after)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidDateRange))
		return
	}
	limit, ok := parseLimit(c, 10)
	if !ok {
		return
	}

	results, err := server.store.GetTransactionsByConnectionType(c.Request.Context(), database.GetTransactionsByConnectionTypeParams{
		ConnectionType: database.NullConnectionType{ConnectionType: connType, Valid: true},
		TransactionAt:  pgtype.Timestamptz{Time: t.UTC(), Valid: true},
		Limit:          int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, results)
}

func (server *Server) getLargeTransactions(c *gin.Context) {
	minAmount := c.Query("min_amount")
	if minAmount == "" {
		c.JSON(http.StatusBadRequest, errorResponse(errMissingQuery))
		return
	}
	var amount pgtype.Numeric
	if err := amount.Scan(minAmount); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidAmount))
		return
	}
	limit, ok := parseLimit(c, 10)
	if !ok {
		return
	}
	results, err := server.store.GetLargeTransactions(c.Request.Context(), database.GetLargeTransactionsParams{
		Amount: amount,
		Limit:  int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, results)
}

func (server *Server) getTransactionsByMetadata(c *gin.Context) {
	raw := c.Query("metadata")
	if raw == "" {
		c.JSON(http.StatusBadRequest, errorResponse(errMissingQuery))
		return
	}
	var payload json.RawMessage
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidMetadataPayload))
		return
	}
	limit, offset, ok := parseLimitOffset(c)
	if !ok {
		return
	}
	results, err := server.store.GetTransactionsByMetadata(c.Request.Context(), database.GetTransactionsByMetadataParams{
		Column1: payload,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, results)
}

func (server *Server) getTransactionByID(c *gin.Context) {
	id := c.Param("id")
	txID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidTransactionID))
		return
	}
	transaction, err := server.store.GetTransactionByID(c.Request.Context(), txID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errTransactionNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, transaction)
}

func (server *Server) getTransactionWithWallets(c *gin.Context) {
	id := c.Param("id")
	txID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidTransactionID))
		return
	}
	transaction, err := server.store.GetTransactionWithWallets(c.Request.Context(), txID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errTransactionNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, transaction)
}

type updateTransactionStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (server *Server) updateTransactionStatus(c *gin.Context) {
	var req updateTransactionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	status := database.TransactionStatus(req.Status)
	if !status.Valid() {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidStatus))
		return
	}
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidTransactionID))
		return
	}
	transaction, err := server.store.UpdateTransactionStatus(c.Request.Context(), database.UpdateTransactionStatusParams{
		ID:     txID,
		Status: status,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errTransactionNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, transaction)
}

func (server *Server) confirmTransaction(c *gin.Context) {
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidTransactionID))
		return
	}
	transaction, err := server.store.ConfirmTransaction(c.Request.Context(), txID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errTransactionNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, transaction)
}

func (server *Server) settingTransaction(c *gin.Context) {
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidTransactionID))
		return
	}
	transaction, err := server.store.SettingTransaction(c.Request.Context(), txID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errTransactionNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, transaction)
}

func (server *Server) settledTransaction(c *gin.Context) {
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidTransactionID))
		return
	}
	transaction, err := server.store.SettledTransaction(c.Request.Context(), txID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errTransactionNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, transaction)
}

func (server *Server) markTransactionSettled(c *gin.Context) {
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidTransactionID))
		return
	}
	transaction, err := server.store.MarkTransactionSettled(c.Request.Context(), txID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errTransactionNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, transaction)
}

func (server *Server) failTransaction(c *gin.Context) {
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidTransactionID))
		return
	}
	if err := server.store.FailTransaction(c.Request.Context(), txID); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "transaction failed"})
}

func (server *Server) listTransactionsByWallet(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	limit, offset, ok := parseLimitOffset(c)
	if !ok {
		return
	}
	results, err := server.store.ListTransactionsByWallet(c.Request.Context(), database.ListTransactionsByWalletParams{
		FromWalletID: walletID,
		Limit:        int32(limit),
		Offset:       int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, results)
}

func (server *Server) listSentTransactions(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	limit, offset, ok := parseLimitOffset(c)
	if !ok {
		return
	}
	results, err := server.store.ListSentTransactions(c.Request.Context(), database.ListSentTransactionsParams{
		FromWalletID: walletID,
		Limit:        int32(limit),
		Offset:       int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, results)
}

func (server *Server) listReceivedTransactions(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	limit, offset, ok := parseLimitOffset(c)
	if !ok {
		return
	}
	results, err := server.store.ListReceivedTransactions(c.Request.Context(), database.ListReceivedTransactionsParams{
		ToWalletID: walletID,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, results)
}

func (server *Server) listPendingTransactions(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	limit, ok := parseLimit(c, 10)
	if !ok {
		return
	}
	results, err := server.store.ListPendingTransactions(c.Request.Context(), database.ListPendingTransactionsParams{
		FromWalletID: walletID,
		Limit:        int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, results)
}

func (server *Server) getTransactionsByDateRange(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	start := c.Query("start")
	end := c.Query("end")
	if start == "" || end == "" {
		c.JSON(http.StatusBadRequest, errorResponse(errMissingQuery))
		return
	}
	startTime, err := time.Parse(time.RFC3339, start)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidDateRange))
		return
	}
	endTime, err := time.Parse(time.RFC3339, end)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidDateRange))
		return
	}
	results, err := server.store.GetTransactionsByDateRange(c.Request.Context(), database.GetTransactionsByDateRangeParams{
		FromWalletID:    walletID,
		TransactionAt:   pgtype.Timestamptz{Time: startTime.UTC(), Valid: true},
		TransactionAt_2: pgtype.Timestamptz{Time: endTime.UTC(), Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, results)
}

func (server *Server) getTransactionStats(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	stats, err := server.store.GetTransactionStats(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (server *Server) getDailyTransactionSummary(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	results, err := server.store.GetDailyTransactionSummary(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, results)
}

func (server *Server) countTransactionsByWallet(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	count, err := server.store.CountTransactionsByWallet(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (server *Server) checkNonceExists(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	nonce, err := strconv.ParseInt(c.Param("nonce"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidNonce))
		return
	}
	exists, err := server.store.CheckNonceExists(c.Request.Context(), database.CheckNonceExistsParams{
		FromWalletID: walletID,
		Nonce:        nonce,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"exists": exists})
}

func parseLimitOffset(c *gin.Context) (int, int, bool) {
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidLimit))
		return 0, 0, false
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil || offsetInt < 0 {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidOffset))
		return 0, 0, false
	}
	return limitInt, offsetInt, true
}

func parseLimit(c *gin.Context, defaultLimit int) (int, bool) {
	limit := c.DefaultQuery("limit", strconv.Itoa(defaultLimit))
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidLimit))
		return 0, false
	}
	return limitInt, true
}
