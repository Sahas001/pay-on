// Package api provides API functionalities for wallet management.
package api

import (
	"net/http"
	"strconv"
	"time"

	database "github.com/Sahas001/pay-on/internal/database/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type createWalletRequest struct {
	Name        string `json:"name" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Pin         string `json:"pin" binding:"required"`
	DeviceID    string `json:"device_id"`
	PublicKey   string `json:"public_key" binding:"required"`
	PrivateKey  string `json:"private_key"`
}

func (server *Server) createWallet(c *gin.Context) {
	var req createWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	pinHash, err := bcrypt.GenerateFromPassword([]byte(req.Pin), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var balance pgtype.Numeric
	if err := balance.Scan("0"); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var deviceID *string
	if req.DeviceID != "" {
		deviceID = &req.DeviceID
	}

	privateKey := req.PrivateKey
	if privateKey == "" {
		privateKey = "client-managed"
	}

	wallet, err := server.store.CreateWallet(c.Request.Context(), database.CreateWalletParams{
		PublicKey:   req.PublicKey,
		PrivateKey:  privateKey,
		Balance:     balance,
		PhoneNumber: req.PhoneNumber,
		Name:        req.Name,
		PinHash:     string(pinHash),
		DeviceID:    deviceID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	createdAt := time.Time{}
	if wallet.CreatedAt.Valid {
		createdAt = wallet.CreatedAt.Time
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":           wallet.ID.String(),
		"public_key":   wallet.PublicKey,
		"balance":      wallet.Balance,
		"phone_number": wallet.PhoneNumber,
		"name":         wallet.Name,
		"is_active":    wallet.IsActive != nil && *wallet.IsActive,
		"device_id":    wallet.DeviceID,
		"created_at":   createdAt,
	})
}

func (server *Server) listWallets(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil || offsetInt < 0 {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := database.ListWalletsParams{
		Limit:  int32(limitInt),
		Offset: int32(offsetInt),
	}

	wallets, err := server.store.ListWallets(c.Request.Context(), arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, wallets)
}

func (server *Server) listActiveWallets(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil || offsetInt < 0 {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := database.ListActiveWalletsParams{
		Limit:  int32(limitInt),
		Offset: int32(offsetInt),
	}

	wallets, err := server.store.ListActiveWallets(c.Request.Context(), arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, wallets)
}

func (server *Server) countWallets(c *gin.Context) {
	count, err := server.store.CountWallets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
	}
	c.JSON(http.StatusOK, gin.H{"Total Wallets": count})
}

func (server *Server) getWalletsNeedingSync(c *gin.Context)  { server.notImplemented(c) }
func (server *Server) searchWalletsByName(c *gin.Context)    { server.notImplemented(c) }
func (server *Server) searchWalletsByPhone(c *gin.Context)   { server.notImplemented(c) }
func (server *Server) getWalletByPhoneNumber(c *gin.Context) { server.notImplemented(c) }
func (server *Server) getWalletByPublicKey(c *gin.Context)   { server.notImplemented(c) }

func (server *Server) getWalletByDeviceID(c *gin.Context) {
	deviceID := c.Param("device_id")
	wallet, err := server.store.GetWalletByDeviceID(c.Request.Context(), &deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, wallet)
}

func (server *Server) getWalletByID(c *gin.Context) {
	id := c.Param("id")
	wallet, err := server.store.GetWalletByID(c.Request.Context(), uuid.MustParse(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, wallet)
}

func (server *Server) getWalletWithBalance(c *gin.Context) {
	id := c.Param("id")
	wallet, err := server.store.GetWalletWithBalance(c.Request.Context(), uuid.MustParse(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	response := database.GetWalletWithBalanceRow{
		ID:          wallet.ID,
		Name:        wallet.Name,
		PhoneNumber: wallet.PhoneNumber,
		Balance:     wallet.Balance,
		IsActive:    wallet.IsActive,
		CreatedAt:   wallet.CreatedAt,
	}
	c.JSON(http.StatusOK, response)
}

func (server *Server) getWalletBalance(c *gin.Context) {
	id := c.Param("id")
	balance, err := server.store.GetWalletBalance(c.Request.Context(), uuid.MustParse(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"balance": balance})
}

func (server *Server) getWalletBalanceHistory(c *gin.Context) {
	walletID := c.Param("id")
	limit := c.DefaultQuery("limit", "10")

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := database.GetWalletBalanceHistoryParams{
		FromWalletID: uuid.MustParse(walletID),
		Limit:        int32(limitInt),
	}

	history, err := server.store.GetWalletBalanceHistory(c.Request.Context(), arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, history)
}

func (server *Server) getWalletDashboard(c *gin.Context)  { server.notImplemented(c) }
func (server *Server) updateWallet(c *gin.Context)        { server.notImplemented(c) }
func (server *Server) updateWalletBalance(c *gin.Context) { server.notImplemented(c) }

type incrementWalletBalanceRequest struct {
	Amount pgtype.Numeric `json:"amount" binding:"required"`
}

func (server *Server) incrementWalletBalance(c *gin.Context) {
	var req incrementWalletBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	id := c.Param("id")
	arg := database.IncrementWalletBalanceParams{
		ID:      uuid.MustParse(id),
		Balance: req.Amount,
	}
	_, err := server.store.IncrementWalletBalance(c.Request.Context(), arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "wallet balance incremented successfully"})
}

type decrementWalletBalanceRequest struct {
	Amount pgtype.Numeric `json:"amount" binding:"required"`
}

func (server *Server) decrementWalletBalance(c *gin.Context) {
	var req decrementWalletBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	id := c.Param("id")
	arg := database.DecrementWalletBalanceParams{
		ID:      uuid.MustParse(id),
		Balance: req.Amount,
	}
	_, err := server.store.DecrementWalletBalance(c.Request.Context(), arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "wallet balance decremented successfully"})
}

type updateWalletPINRequest struct {
	PIN string `json:"pin" binding:"required"`
}

func (server *Server) updateWalletPIN(c *gin.Context) {
	var req updateWalletPINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	id := c.Param("id")
	pinHash, err := bcrypt.GenerateFromPassword([]byte(req.PIN), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	arg := database.UpdateWalletPINParams{
		ID:      uuid.MustParse(id),
		PinHash: string(pinHash),
	}
	err = server.store.UpdateWalletPIN(c.Request.Context(), arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "wallet PIN updated successfully"})
}
func (server *Server) updateWalletLastSync(c *gin.Context) { server.notImplemented(c) }
func (server *Server) deactivateWallet(c *gin.Context) {
	id := c.Param("id")
	err := server.store.DeactivateWallet(c.Request.Context(), uuid.MustParse(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
	}
	c.JSON(http.StatusOK, gin.H{"message": "wallet deactivated successfully"})
}

func (server *Server) activateWallet(c *gin.Context) {
	id := c.Param("id")
	err := server.store.ActivateWallet(c.Request.Context(), uuid.MustParse(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
	}
	c.JSON(http.StatusOK, gin.H{"message": "wallet activated successfully"})
}

func (server *Server) softDeleteWallet(c *gin.Context) {
	id := c.Param("id")
	err := server.store.SoftDeleteWallet(c.Request.Context(), uuid.MustParse(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "wallet soft deleted successfully"})
}

func (server *Server) hardDeleteWallet(c *gin.Context) {
	id := c.Param("id")
	err := server.store.HardDeleteWallet(c.Request.Context(), uuid.MustParse(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "wallet hard deleted successfully"})
}
