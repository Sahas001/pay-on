package api

import "github.com/gin-gonic/gin"

func (server *Server) createPeer(c *gin.Context)             { server.notImplemented(c) }
func (server *Server) upsertPeer(c *gin.Context)             { server.notImplemented(c) }
func (server *Server) autoTrustFrequentPeers(c *gin.Context) { server.notImplemented(c) }
func (server *Server) getPeerByID(c *gin.Context)            { server.notImplemented(c) }
func (server *Server) updatePeerLastSeen(c *gin.Context)     { server.notImplemented(c) }
func (server *Server) setPeerTrusted(c *gin.Context)         { server.notImplemented(c) }
func (server *Server) deletePeer(c *gin.Context)             { server.notImplemented(c) }
func (server *Server) hardDeletePeer(c *gin.Context)         { server.notImplemented(c) }
func (server *Server) listPeersByWallet(c *gin.Context)      { server.notImplemented(c) }
func (server *Server) listTrustedPeers(c *gin.Context)       { server.notImplemented(c) }
func (server *Server) listRecentPeers(c *gin.Context)        { server.notImplemented(c) }
func (server *Server) listPeersByConnectionType(c *gin.Context) {
	server.notImplemented(c)
}
func (server *Server) getTopPeersByVolume(c *gin.Context) { server.notImplemented(c) }
func (server *Server) getTopPeersByTransactionCount(c *gin.Context) {
	server.notImplemented(c)
}
func (server *Server) countPeersByWallet(c *gin.Context)       { server.notImplemented(c) }
func (server *Server) countTrustedPeers(c *gin.Context)        { server.notImplemented(c) }
func (server *Server) getStalePeers(c *gin.Context)            { server.notImplemented(c) }
func (server *Server) getPeerByWalletAndPeerID(c *gin.Context) { server.notImplemented(c) }
func (server *Server) updatePeerInfo(c *gin.Context)           { server.notImplemented(c) }
func (server *Server) incrementPeerTransactionCount(c *gin.Context) {
	server.notImplemented(c)
}
