// Package api provides API functionalities for wallet management.
package api

import "github.com/gin-gonic/gin"

func (server *Server) createWallet(c *gin.Context)            { server.notImplemented(c) }
func (server *Server) listWallets(c *gin.Context)             { server.notImplemented(c) }
func (server *Server) listActiveWallets(c *gin.Context)       { server.notImplemented(c) }
func (server *Server) countWallets(c *gin.Context)            { server.notImplemented(c) }
func (server *Server) getWalletsNeedingSync(c *gin.Context)   { server.notImplemented(c) }
func (server *Server) searchWalletsByName(c *gin.Context)     { server.notImplemented(c) }
func (server *Server) searchWalletsByPhone(c *gin.Context)    { server.notImplemented(c) }
func (server *Server) getWalletByPhoneNumber(c *gin.Context)  { server.notImplemented(c) }
func (server *Server) getWalletByPublicKey(c *gin.Context)    { server.notImplemented(c) }
func (server *Server) getWalletByDeviceID(c *gin.Context)     { server.notImplemented(c) }
func (server *Server) getWalletByID(c *gin.Context)           { server.notImplemented(c) }
func (server *Server) getWalletWithBalance(c *gin.Context)    { server.notImplemented(c) }
func (server *Server) getWalletBalance(c *gin.Context)        { server.notImplemented(c) }
func (server *Server) getWalletBalanceHistory(c *gin.Context) { server.notImplemented(c) }
func (server *Server) getWalletDashboard(c *gin.Context)      { server.notImplemented(c) }
func (server *Server) updateWallet(c *gin.Context)            { server.notImplemented(c) }
func (server *Server) updateWalletBalance(c *gin.Context)     { server.notImplemented(c) }
func (server *Server) incrementWalletBalance(c *gin.Context)  { server.notImplemented(c) }
func (server *Server) decrementWalletBalance(c *gin.Context)  { server.notImplemented(c) }
func (server *Server) updateWalletPIN(c *gin.Context)         { server.notImplemented(c) }
func (server *Server) updateWalletLastSync(c *gin.Context)    { server.notImplemented(c) }
func (server *Server) deactivateWallet(c *gin.Context)        { server.notImplemented(c) }
func (server *Server) activateWallet(c *gin.Context)          { server.notImplemented(c) }
func (server *Server) softDeleteWallet(c *gin.Context)        { server.notImplemented(c) }
func (server *Server) hardDeleteWallet(c *gin.Context)        { server.notImplemented(c) }
