package api

import "github.com/gin-gonic/gin"

func (server *Server) createSyncLog(c *gin.Context)           { server.notImplemented(c) }
func (server *Server) listAllPendingSyncs(c *gin.Context)     { server.notImplemented(c) }
func (server *Server) getSyncsNeedingRetry(c *gin.Context)    { server.notImplemented(c) }
func (server *Server) countSyncLogsByStatus(c *gin.Context)   { server.notImplemented(c) }
func (server *Server) deleteOldSyncLogs(c *gin.Context)       { server.notImplemented(c) }
func (server *Server) getSyncLogByID(c *gin.Context)          { server.notImplemented(c) }
func (server *Server) updateSyncLogStatus(c *gin.Context)     { server.notImplemented(c) }
func (server *Server) markSettleSuccessful(c *gin.Context)    { server.notImplemented(c) }
func (server *Server) markSettleFailed(c *gin.Context)        { server.notImplemented(c) }
func (server *Server) markSettleConflict(c *gin.Context)      { server.notImplemented(c) }
func (server *Server) resolveSyncConflict(c *gin.Context)     { server.notImplemented(c) }
func (server *Server) getSyncLogsByWallet(c *gin.Context)     { server.notImplemented(c) }
func (server *Server) listPendingSyncs(c *gin.Context)        { server.notImplemented(c) }
func (server *Server) listFailedSyncs(c *gin.Context)         { server.notImplemented(c) }
func (server *Server) listConflictedSyncs(c *gin.Context)     { server.notImplemented(c) }
func (server *Server) getSyncStats(c *gin.Context)            { server.notImplemented(c) }
func (server *Server) getSyncLogsByTransaction(c *gin.Context) { server.notImplemented(c) }
