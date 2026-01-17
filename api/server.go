package api

import (
	database "github.com/Sahas001/pay-on/internal/database/sqlc"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	store  *database.Store
	router *gin.Engine
}

func NewServer(store *database.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			if origin == "" || origin == "null" {
				return true
			}
			switch origin {
			case "http://localhost:3000", "http://localhost:5173", "http://localhost:8080", "http://127.0.0.1:5500":
				return true
			default:
				return false
			}
		},
		AllowMethods: []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	wallets := router.Group("/wallets")
	wallets.POST("", server.createWallet)
	wallets.GET("", server.listWallets)
	wallets.GET("/active", server.listActiveWallets)
	wallets.GET("/count", server.countWallets)
	wallets.GET("/needs-sync", server.getWalletsNeedingSync)
	wallets.GET("/search/name", server.searchWalletsByName)
	wallets.GET("/search/phone", server.searchWalletsByPhone)
	wallets.GET("/phone/:phone", server.getWalletByPhoneNumber)
	wallets.GET("/public/:public_key", server.getWalletByPublicKey)
	wallets.GET("/device/:device_id", server.getWalletByDeviceID)
	wallets.GET("/:id", server.getWalletByID)
	wallets.GET("/:id/summary", server.getWalletWithBalance)
	wallets.GET("/:id/balance", server.getWalletBalance)
	wallets.GET("/:id/balance-history", server.getWalletBalanceHistory)
	wallets.GET("/:id/dashboard", server.getWalletDashboard)
	wallets.PATCH("/:id", server.updateWallet)
	wallets.PATCH("/:id/balance", server.updateWalletBalance)
	wallets.POST("/:id/balance/increment", server.incrementWalletBalance)
	wallets.POST("/:id/balance/decrement", server.decrementWalletBalance)
	wallets.PATCH("/:id/pin", server.updateWalletPIN)
	wallets.PATCH("/:id/sync", server.updateWalletLastSync)
	wallets.POST("/:id/deactivate", server.deactivateWallet)
	wallets.POST("/:id/activate", server.activateWallet)
	wallets.DELETE("/:id", server.softDeleteWallet)
	wallets.DELETE("/:id/hard", server.hardDeleteWallet)

	transactions := router.Group("/transactions")
	transactions.POST("", server.createTransaction)
	transactions.GET("/search", server.searchTransactions)
	transactions.GET("/recent", server.getRecentTransactions)
	transactions.GET("/status/:status", server.listTransactionsByStatus)
	transactions.GET("/unsynced", server.listUnsyncedTransactions)
	transactions.GET("/pending/count", server.countPendingTransactions)
	transactions.GET("/connection/:type", server.getTransactionsByConnectionType)
	transactions.GET("/large", server.getLargeTransactions)
	transactions.GET("/metadata", server.getTransactionsByMetadata)
	transactions.GET("/:id", server.getTransactionByID)
	transactions.GET("/:id/with-wallets", server.getTransactionWithWallets)
	transactions.PATCH("/:id/status", server.updateTransactionStatus)
	transactions.POST("/:id/confirm", server.confirmTransaction)
	transactions.POST("/:id/settling", server.settingTransaction)
	transactions.POST("/:id/settled", server.settledTransaction)
	transactions.POST("/:id/mark-settled", server.markTransactionSettled)
	transactions.POST("/:id/fail", server.failTransaction)

	walletTransactions := router.Group("/wallets/:id/transactions")
	walletTransactions.GET("", server.listTransactionsByWallet)
	walletTransactions.GET("/sent", server.listSentTransactions)
	walletTransactions.GET("/received", server.listReceivedTransactions)
	walletTransactions.GET("/pending", server.listPendingTransactions)
	walletTransactions.GET("/range", server.getTransactionsByDateRange)
	walletTransactions.GET("/stats", server.getTransactionStats)
	walletTransactions.GET("/daily-summary", server.getDailyTransactionSummary)
	walletTransactions.GET("/count", server.countTransactionsByWallet)
	walletTransactions.GET("/nonce/:nonce", server.checkNonceExists)

	peers := router.Group("/peers")
	peers.POST("", server.createPeer)
	peers.POST("/upsert", server.upsertPeer)
	peers.POST("/auto-trust", server.autoTrustFrequentPeers)
	peers.GET("/:id", server.getPeerByID)
	peers.PATCH("/:id/last-seen", server.updatePeerLastSeen)
	peers.DELETE("/:id", server.deletePeer)
	peers.DELETE("/:id/hard", server.hardDeletePeer)

	walletPeers := router.Group("/wallets/:id/peers")
	walletPeers.GET("", server.listPeersByWallet)
	walletPeers.GET("/trusted", server.listTrustedPeers)
	walletPeers.GET("/recent", server.listRecentPeers)
	walletPeers.GET("/connection/:type", server.listPeersByConnectionType)
	walletPeers.GET("/top-volume", server.getTopPeersByVolume)
	walletPeers.GET("/top-count", server.getTopPeersByTransactionCount)
	walletPeers.GET("/count", server.countPeersByWallet)
	walletPeers.GET("/count/trusted", server.countTrustedPeers)
	walletPeers.GET("/stale", server.getStalePeers)
	walletPeers.GET("/:peer_id", server.getPeerByWalletAndPeerID)
	walletPeers.PATCH("/:peer_id", server.updatePeerInfo)
	walletPeers.PATCH("/:peer_id/trusted", server.setPeerTrustedByWallet)
	walletPeers.POST("/:peer_id/transaction-count", server.incrementPeerTransactionCount)

	syncLogs := router.Group("/sync-logs")
	syncLogs.POST("", server.createSyncLog)
	syncLogs.GET("/pending", server.listAllPendingSyncs)
	syncLogs.GET("/retry", server.getSyncsNeedingRetry)
	syncLogs.GET("/count/:status", server.countSyncLogsByStatus)
	syncLogs.DELETE("/old", server.deleteOldSyncLogs)
	syncLogs.GET("/:id", server.getSyncLogByID)
	syncLogs.PATCH("/:id/status", server.updateSyncLogStatus)
	syncLogs.POST("/:id/settle-success", server.markSettleSuccessful)
	syncLogs.POST("/:id/settle-failed", server.markSettleFailed)
	syncLogs.POST("/:id/settle-conflict", server.markSettleConflict)
	syncLogs.POST("/:id/resolve", server.resolveSyncConflict)

	walletSyncLogs := router.Group("/wallets/:id/sync-logs")
	walletSyncLogs.GET("", server.getSyncLogsByWallet)
	walletSyncLogs.GET("/pending", server.listPendingSyncs)
	walletSyncLogs.GET("/failed", server.listFailedSyncs)
	walletSyncLogs.GET("/conflicts", server.listConflictedSyncs)
	walletSyncLogs.GET("/stats", server.getSyncStats)

	transactionSyncLogs := router.Group("/transactions/:id/sync-logs")
	transactionSyncLogs.GET("", server.getSyncLogsByTransaction)

	auditLogs := router.Group("/audit-logs")
	auditLogs.POST("", server.createAuditLog)
	auditLogs.GET("", server.listAuditLogs)
	auditLogs.GET("/recent", server.getRecentAuditLogs)
	auditLogs.GET("/count", server.countAuditLogs)
	auditLogs.GET("/count/:table", server.countAuditLogsByTable)
	auditLogs.GET("/table/:table", server.listAuditLogsByTable)
	auditLogs.GET("/record/:table/:record_id", server.listAuditLogsByRecord)
	auditLogs.GET("/action/:action", server.listAuditLogsByAction)
	auditLogs.GET("/user/:user_id", server.listAuditLogsByUser)
	auditLogs.GET("/range", server.listAuditLogsByDateRange)
	auditLogs.GET("/ip/:ip", server.listAuditLogsByIP)
	auditLogs.GET("/history/:table/:record_id", server.getRecordHistory)
	auditLogs.GET("/balance-history/:wallet_id", server.getBalanceHistory)
	auditLogs.DELETE("/old", server.deleteOldAuditLogs)
	auditLogs.GET("/:id", server.getAuditLogByID)

	stats := router.Group("/stats")
	stats.GET("/system", server.getSystemStats)

	router.POST("/transfers", server.transferTx)

	server.router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func okayResponse(message string) gin.H {
	return gin.H{"message": message}
}
