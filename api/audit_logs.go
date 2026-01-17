package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/netip"
	"strconv"
	"time"

	database "github.com/Sahas001/pay-on/internal/database/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	errInvalidAuditLogID = errors.New("invalid audit log id")
	errInvalidRecordID   = errors.New("invalid record id")
	errInvalidUserID     = errors.New("invalid user id")
	errInvalidIPAddress  = errors.New("invalid ip address")
	errInvalidDateRange  = errors.New("invalid date range")
	errInvalidDays       = errors.New("invalid days")
	errAuditLogNotFound  = errors.New("audit log not found")
	errInvalidMetadata   = errors.New("invalid json payload")
	errInvalidTableName  = errors.New("invalid table name")
	errInvalidAction     = errors.New("invalid action")
)

type createAuditLogRequest struct {
	TableName string          `json:"table_name" binding:"required"`
	RecordID  uuid.UUID       `json:"record_id" binding:"required"`
	Action    string          `json:"action" binding:"required"`
	OldData   json.RawMessage `json:"old_data"`
	NewData   json.RawMessage `json:"new_data"`
	ChangedBy string          `json:"changed_by"`
	IPAddress string          `json:"ip_address"`
	UserAgent *string         `json:"user_agent"`
}

func (server *Server) createAuditLog(c *gin.Context) {
	var req createAuditLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if req.TableName == "" {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidTableName))
		return
	}
	if req.Action == "" {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidAction))
		return
	}

	var changedBy pgtype.UUID
	if req.ChangedBy != "" {
		parsed, err := uuid.Parse(req.ChangedBy)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse(errInvalidUserID))
			return
		}
		copy(changedBy.Bytes[:], parsed[:])
		changedBy.Valid = true
	}

	var ip *netip.Addr
	if req.IPAddress != "" {
		parsed, err := netip.ParseAddr(req.IPAddress)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse(errInvalidIPAddress))
			return
		}
		ip = &parsed
	}

	oldData := req.OldData
	if len(oldData) == 0 {
		oldData = []byte(`{}`)
	}
	newData := req.NewData
	if len(newData) == 0 {
		newData = []byte(`{}`)
	}

	log, err := server.store.CreateAuditLog(c.Request.Context(), database.CreateAuditLogParams{
		TableName: req.TableName,
		RecordID:  req.RecordID,
		Action:    req.Action,
		OldData:   oldData,
		NewData:   newData,
		ChangedBy: changedBy,
		IpAddress: ip,
		UserAgent: req.UserAgent,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusCreated, log)
}

func (server *Server) listAuditLogs(c *gin.Context) {
	limit, offset, ok := parseLimitOffset(c)
	if !ok {
		return
	}
	logs, err := server.store.ListAuditLogs(c.Request.Context(), database.ListAuditLogsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (server *Server) getRecentAuditLogs(c *gin.Context) {
	limit, ok := parseLimit(c, 10)
	if !ok {
		return
	}
	logs, err := server.store.GetRecentAuditLogs(c.Request.Context(), int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (server *Server) countAuditLogs(c *gin.Context) {
	count, err := server.store.CountAuditLogs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (server *Server) countAuditLogsByTable(c *gin.Context) {
	table := c.Param("table")
	if table == "" {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidTableName))
		return
	}
	count, err := server.store.CountAuditLogsByTable(c.Request.Context(), table)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (server *Server) listAuditLogsByTable(c *gin.Context) {
	table := c.Param("table")
	if table == "" {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidTableName))
		return
	}
	limit, offset, ok := parseLimitOffset(c)
	if !ok {
		return
	}
	logs, err := server.store.ListAuditLogsByTable(c.Request.Context(), database.ListAuditLogsByTableParams{
		TableName: table,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (server *Server) listAuditLogsByRecord(c *gin.Context) {
	table := c.Param("table")
	recordID, err := uuid.Parse(c.Param("record_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidRecordID))
		return
	}
	logs, err := server.store.ListAuditLogsByRecord(c.Request.Context(), database.ListAuditLogsByRecordParams{
		TableName: table,
		RecordID:  recordID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (server *Server) listAuditLogsByAction(c *gin.Context) {
	action := c.Param("action")
	if action == "" {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidAction))
		return
	}
	limit, offset, ok := parseLimitOffset(c)
	if !ok {
		return
	}
	logs, err := server.store.ListAuditLogsByAction(c.Request.Context(), database.ListAuditLogsByActionParams{
		Action: action,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (server *Server) listAuditLogsByUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidUserID))
		return
	}
	limit, offset, ok := parseLimitOffset(c)
	if !ok {
		return
	}
	var changedBy pgtype.UUID
	copy(changedBy.Bytes[:], userID[:])
	changedBy.Valid = true

	logs, err := server.store.ListAuditLogsByUser(c.Request.Context(), database.ListAuditLogsByUserParams{
		ChangedBy: changedBy,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (server *Server) listAuditLogsByDateRange(c *gin.Context) {
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
	limit, offset, ok := parseLimitOffset(c)
	if !ok {
		return
	}
	logs, err := server.store.ListAuditLogsByDateRange(c.Request.Context(), database.ListAuditLogsByDateRangeParams{
		ChangedAt:   pgtype.Timestamptz{Time: startTime.UTC(), Valid: true},
		ChangedAt_2: pgtype.Timestamptz{Time: endTime.UTC(), Valid: true},
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (server *Server) listAuditLogsByIP(c *gin.Context) {
	raw := c.Param("ip")
	addr, err := netip.ParseAddr(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidIPAddress))
		return
	}
	limit, offset, ok := parseLimitOffset(c)
	if !ok {
		return
	}
	logs, err := server.store.ListAuditLogsByIP(c.Request.Context(), database.ListAuditLogsByIPParams{
		IpAddress: &addr,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (server *Server) getRecordHistory(c *gin.Context) {
	table := c.Param("table")
	recordID, err := uuid.Parse(c.Param("record_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidRecordID))
		return
	}
	logs, err := server.store.GetRecordHistory(c.Request.Context(), database.GetRecordHistoryParams{
		TableName: table,
		RecordID:  recordID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (server *Server) getBalanceHistory(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("wallet_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	limit, ok := parseLimit(c, 10)
	if !ok {
		return
	}
	history, err := server.store.GetBalanceHistory(c.Request.Context(), database.GetBalanceHistoryParams{
		RecordID: walletID,
		Limit:    int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, history)
}

func (server *Server) deleteOldAuditLogs(c *gin.Context) {
	days := c.Query("days")
	if days == "" {
		c.JSON(http.StatusBadRequest, errorResponse(errMissingQuery))
		return
	}
	if _, err := strconv.Atoi(days); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidDays))
		return
	}
	if err := server.store.DeleteOldAuditLogs(c.Request.Context(), &days); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "audit logs deleted"})
}

func (server *Server) getAuditLogByID(c *gin.Context) {
	id := c.Param("id")
	logID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidAuditLogID))
		return
	}
	log, err := server.store.GetAuditLogByID(c.Request.Context(), logID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errAuditLogNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, log)
}
