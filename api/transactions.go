package api

import "github.com/gin-gonic/gin"

func (server *Server) createTransaction(c *gin.Context)        { server.notImplemented(c) }
func (server *Server) searchTransactions(c *gin.Context)       { server.notImplemented(c) }
func (server *Server) getRecentTransactions(c *gin.Context)    { server.notImplemented(c) }
func (server *Server) listTransactionsByStatus(c *gin.Context) { server.notImplemented(c) }
func (server *Server) listUnsyncedTransactions(c *gin.Context) { server.notImplemented(c) }
func (server *Server) countPendingTransactions(c *gin.Context) { server.notImplemented(c) }
func (server *Server) getTransactionsByConnectionType(c *gin.Context) {
	server.notImplemented(c)
}
func (server *Server) getLargeTransactions(c *gin.Context)      { server.notImplemented(c) }
func (server *Server) getTransactionsByMetadata(c *gin.Context) { server.notImplemented(c) }
func (server *Server) getTransactionByID(c *gin.Context)        { server.notImplemented(c) }
func (server *Server) getTransactionWithWallets(c *gin.Context) { server.notImplemented(c) }
func (server *Server) updateTransactionStatus(c *gin.Context)   { server.notImplemented(c) }
func (server *Server) confirmTransaction(c *gin.Context)        { server.notImplemented(c) }
func (server *Server) settingTransaction(c *gin.Context)        { server.notImplemented(c) }
func (server *Server) settledTransaction(c *gin.Context)        { server.notImplemented(c) }
func (server *Server) markTransactionSettled(c *gin.Context)    { server.notImplemented(c) }
func (server *Server) failTransaction(c *gin.Context)           { server.notImplemented(c) }
func (server *Server) listTransactionsByWallet(c *gin.Context)  { server.notImplemented(c) }
func (server *Server) listSentTransactions(c *gin.Context)      { server.notImplemented(c) }
func (server *Server) listReceivedTransactions(c *gin.Context)  { server.notImplemented(c) }
func (server *Server) listPendingTransactions(c *gin.Context)   { server.notImplemented(c) }
func (server *Server) getTransactionsByDateRange(c *gin.Context) {
	server.notImplemented(c)
}
func (server *Server) getTransactionStats(c *gin.Context)        { server.notImplemented(c) }
func (server *Server) getDailyTransactionSummary(c *gin.Context) { server.notImplemented(c) }
func (server *Server) countTransactionsByWallet(c *gin.Context)  { server.notImplemented(c) }
func (server *Server) checkNonceExists(c *gin.Context)           { server.notImplemented(c) }
