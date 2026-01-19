package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	database "github.com/Sahas001/pay-on/internal/database/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

var (
	errInvalidTransferType   = errors.New("invalid transfer type")
	errInvalidTransferStatus = errors.New("invalid transfer status")
)

type transferRequest struct {
	FromWalletID   string          `json:"from_wallet_id" binding:"required"`
	ToWalletID     string          `json:"to_wallet_id" binding:"required"`
	Amount         string          `json:"amount" binding:"required"`
	Pin            string          `json:"pin" binding:"required"`
	Currency       string          `json:"currency"`
	Type           string          `json:"type"`
	Status         string          `json:"status"`
	Signature      string          `json:"signature" binding:"required"`
	Nonce          int64           `json:"nonce"`
	ConnectionType string          `json:"connection_type"`
	Description    *string         `json:"description"`
	Metadata       json.RawMessage `json:"metadata"`
	TransactionAt  *time.Time      `json:"transaction_at"`
}

func (server *Server) transferTx(c *gin.Context) {
	var req transferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	userID, ok := authUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse(errInvalidCredentials))
		return
	}

	fromWalletID, err := uuid.Parse(req.FromWalletID)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	toWalletID, err := uuid.Parse(req.ToWalletID)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}

	var amount pgtype.Numeric
	if err := amount.Scan(req.Amount); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidAmount))
		return
	}

	fromWallet, err := server.store.GetWalletByID(c.Request.Context(), fromWalletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if !fromWallet.UserID.Valid || fromWallet.UserID.Bytes != toPgUUID(userID).Bytes {
		c.JSON(http.StatusUnauthorized, errorResponse(errInvalidCredentials))
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(fromWallet.PinHash), []byte(req.Pin)); err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse(errInvalidCredentials))
		return
	}

	txType := database.TransactionType(req.Type)
	if req.Type != "" && !txType.Valid() {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidTransferType))
		return
	}

	txStatus := database.TransactionStatus(req.Status)
	if req.Status != "" && !txStatus.Valid() {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidTransferStatus))
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

	var txTime pgtype.Timestamptz
	if req.TransactionAt != nil {
		txTime = pgtype.Timestamptz{Time: req.TransactionAt.UTC(), Valid: true}
	}

	result, err := server.store.TransferTx(c.Request.Context(), database.TransferTxParams{
		FromWalletID:   fromWalletID,
		ToWalletID:     toWalletID,
		Amount:         amount,
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

	c.JSON(http.StatusOK, result)
}
