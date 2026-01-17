package api

import (
	"errors"
	"net"
	"net/http"
	"net/netip"

	database "github.com/Sahas001/pay-on/internal/database/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	errInvalidPeerID        = errors.New("invalid peer id")
	errInvalidPeerWalletID  = errors.New("invalid peer wallet id")
	errInvalidConnection    = errors.New("invalid connection type")
	errInvalidMacAddress    = errors.New("invalid bluetooth address")
	errPeerNotFound         = errors.New("peer not found")
	errInvalidTransactionCt = errors.New("invalid transaction count")
)

type createPeerRequest struct {
	WalletID       uuid.UUID `json:"wallet_id" binding:"required"`
	PeerWalletID   uuid.UUID `json:"peer_wallet_id" binding:"required"`
	Name           *string   `json:"name"`
	PublicKey      string    `json:"public_key" binding:"required"`
	IPAddress      string    `json:"ip_address"`
	BluetoothAddr  string    `json:"bt_address"`
	ConnectionType string    `json:"connection_type" binding:"required"`
	IsTrusted      *bool     `json:"is_trusted"`
}

func (server *Server) createPeer(c *gin.Context) {
	var req createPeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	conn := database.ConnectionType(req.ConnectionType)
	if !conn.Valid() {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidConnection))
		return
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

	mac, err := net.ParseMAC(req.BluetoothAddr)
	if err != nil {
		if req.BluetoothAddr != "" {
			c.JSON(http.StatusBadRequest, errorResponse(errInvalidMacAddress))
			return
		}
		mac = net.HardwareAddr{}
	}

	peer, err := server.store.CreatePeer(c.Request.Context(), database.CreatePeerParams{
		WalletID:       req.WalletID,
		PeerWalletID:   req.PeerWalletID,
		Name:           req.Name,
		PublicKey:      req.PublicKey,
		IpAddress:      ip,
		BtAddress:      mac,
		ConnectionType: conn,
		IsTrusted:      req.IsTrusted,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusCreated, peer)
}

func (server *Server) upsertPeer(c *gin.Context) {
	var req createPeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	conn := database.ConnectionType(req.ConnectionType)
	if !conn.Valid() {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidConnection))
		return
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

	mac, err := net.ParseMAC(req.BluetoothAddr)
	if err != nil {
		if req.BluetoothAddr != "" {
			c.JSON(http.StatusBadRequest, errorResponse(errInvalidMacAddress))
			return
		}
		mac = net.HardwareAddr{}
	}

	peer, err := server.store.UpsertPeer(c.Request.Context(), database.UpsertPeerParams{
		WalletID:       req.WalletID,
		PeerWalletID:   req.PeerWalletID,
		Name:           req.Name,
		PublicKey:      req.PublicKey,
		IpAddress:      ip,
		BtAddress:      mac,
		ConnectionType: conn,
		IsTrusted:      req.IsTrusted,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, peer)
}

type autoTrustFrequentPeersRequest struct {
	TransactionCount int32 `json:"transaction_count" binding:"required"`
}

func (server *Server) autoTrustFrequentPeers(c *gin.Context) {
	var req autoTrustFrequentPeersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if req.TransactionCount < 0 {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidTransactionCt))
		return
	}
	count := req.TransactionCount
	if err := server.store.AutoTrustFrequentPeers(c.Request.Context(), &count); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "peers updated"})
}

func (server *Server) getPeerByID(c *gin.Context) {
	id := c.Param("id")
	peerID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidPeerID))
		return
	}
	peer, err := server.store.GetPeerByID(c.Request.Context(), peerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errPeerNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, peer)
}

func (server *Server) updatePeerLastSeen(c *gin.Context) {
	id := c.Param("id")
	peerID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidPeerID))
		return
	}
	if err := server.store.UpdatePeerLastSeen(c.Request.Context(), peerID); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "peer updated"})
}

type setPeerTrustedRequest struct {
	IsTrusted bool `json:"is_trusted" binding:"required"`
}

func (server *Server) deletePeer(c *gin.Context) {
	id := c.Param("id")
	peerID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidPeerID))
		return
	}
	if err := server.store.DeletePeer(c.Request.Context(), peerID); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "peer deleted"})
}

func (server *Server) hardDeletePeer(c *gin.Context) {
	id := c.Param("id")
	peerID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidPeerID))
		return
	}
	if err := server.store.HardDeletePeer(c.Request.Context(), peerID); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "peer deleted"})
}

func (server *Server) listPeersByWallet(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	limit, offset, ok := parseLimitOffset(c)
	if !ok {
		return
	}
	peers, err := server.store.ListPeersByWallet(c.Request.Context(), database.ListPeersByWalletParams{
		WalletID: walletID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, peers)
}

func (server *Server) listTrustedPeers(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	peers, err := server.store.ListTrustedPeers(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, peers)
}

func (server *Server) listRecentPeers(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	limit, ok := parseLimit(c, 10)
	if !ok {
		return
	}
	peers, err := server.store.ListRecentPeers(c.Request.Context(), database.ListRecentPeersParams{
		WalletID: walletID,
		Limit:    int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, peers)
}

func (server *Server) listPeersByConnectionType(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	conn := database.ConnectionType(c.Param("type"))
	if !conn.Valid() {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidConnection))
		return
	}
	limit, ok := parseLimit(c, 10)
	if !ok {
		return
	}
	peers, err := server.store.ListPeersByConnectionType(c.Request.Context(), database.ListPeersByConnectionTypeParams{
		WalletID:       walletID,
		ConnectionType: conn,
		Limit:          int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, peers)
}

func (server *Server) getTopPeersByVolume(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	limit, ok := parseLimit(c, 10)
	if !ok {
		return
	}
	peers, err := server.store.GetTopPeersByVolume(c.Request.Context(), database.GetTopPeersByVolumeParams{
		WalletID: walletID,
		Limit:    int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, peers)
}

func (server *Server) getTopPeersByTransactionCount(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	limit, ok := parseLimit(c, 10)
	if !ok {
		return
	}
	peers, err := server.store.GetTopPeersByTransactionCount(c.Request.Context(), database.GetTopPeersByTransactionCountParams{
		WalletID: walletID,
		Limit:    int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, peers)
}

func (server *Server) countPeersByWallet(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	count, err := server.store.CountPeersByWallet(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (server *Server) countTrustedPeers(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	count, err := server.store.CountTrustedPeers(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (server *Server) getStalePeers(c *gin.Context) {
	limit, ok := parseLimit(c, 10)
	if !ok {
		return
	}
	peers, err := server.store.GetStalePeers(c.Request.Context(), int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, peers)
}

func (server *Server) getPeerByWalletAndPeerID(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	peerWalletID, err := uuid.Parse(c.Param("peer_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidPeerWalletID))
		return
	}
	peer, err := server.store.GetPeerByWalletAndPeerID(c.Request.Context(), database.GetPeerByWalletAndPeerIDParams{
		WalletID:     walletID,
		PeerWalletID: peerWalletID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errPeerNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, peer)
}

type updatePeerInfoRequest struct {
	Name           *string `json:"name"`
	IPAddress      string  `json:"ip_address"`
	BluetoothAddr  string  `json:"bt_address"`
	ConnectionType string  `json:"connection_type"`
}

func (server *Server) updatePeerInfo(c *gin.Context) {
	var req updatePeerInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	peerWalletID, err := uuid.Parse(c.Param("peer_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidPeerWalletID))
		return
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

	mac, err := net.ParseMAC(req.BluetoothAddr)
	if err != nil {
		if req.BluetoothAddr != "" {
			c.JSON(http.StatusBadRequest, errorResponse(errInvalidMacAddress))
			return
		}
		mac = net.HardwareAddr{}
	}

	var conn database.NullConnectionType
	if req.ConnectionType != "" {
		typed := database.ConnectionType(req.ConnectionType)
		if !typed.Valid() {
			c.JSON(http.StatusBadRequest, errorResponse(errInvalidConnection))
			return
		}
		conn = database.NullConnectionType{ConnectionType: typed, Valid: true}
	}

	peer, err := server.store.UpdatePeerInfo(c.Request.Context(), database.UpdatePeerInfoParams{
		WalletID:       walletID,
		PeerWalletID:   peerWalletID,
		Name:           req.Name,
		IpAddress:      ip,
		BtAddress:      mac,
		ConnectionType: conn,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errPeerNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, peer)
}

func (server *Server) incrementPeerTransactionCount(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	peerWalletID, err := uuid.Parse(c.Param("peer_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidPeerWalletID))
		return
	}
	if err := server.store.IncrementPeerTransactionCount(c.Request.Context(), database.IncrementPeerTransactionCountParams{
		WalletID:     walletID,
		PeerWalletID: peerWalletID,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "peer transaction count incremented"})
}

func (server *Server) setPeerTrustedByWallet(c *gin.Context) {
	var req setPeerTrustedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidWalletID))
		return
	}
	peerWalletID, err := uuid.Parse(c.Param("peer_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(errInvalidPeerWalletID))
		return
	}

	peer, err := server.store.GetPeerByWalletAndPeerID(c.Request.Context(), database.GetPeerByWalletAndPeerIDParams{
		WalletID:     walletID,
		PeerWalletID: peerWalletID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(errPeerNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	trusted := req.IsTrusted
	if err := server.store.SetPeerTrusted(c.Request.Context(), database.SetPeerTrustedParams{
		ID:        peer.ID,
		IsTrusted: &trusted,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "peer updated"})
}
