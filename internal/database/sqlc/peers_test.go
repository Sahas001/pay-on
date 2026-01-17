package database

import (
	"context"
	"net"
	"net/netip"
	"testing"
)

func TestPeerQueries(t *testing.T) {
	withTx(t, func(ctx context.Context, q *Queries) {
		wallet := createTestWallet(t, ctx, q)
		peerWallet := createTestWallet(t, ctx, q)

		peer := createTestPeer(t, ctx, q, wallet.ID, peerWallet.ID)

		walletAndPeerID := GetPeerByWalletAndPeerIDParams{
			WalletID:     wallet.ID,
			PeerWalletID: peerWallet.ID,
		}

		fetched, err := q.GetPeerByWalletAndPeerID(ctx, walletAndPeerID)
		if err != nil {
			t.Fatalf("get peer by wallet and peer id: %v", err)
		}
		if fetched.ID != peer.ID {
			t.Fatalf("expected peer id %s, got %s", peer.ID, fetched.ID)
		}

		newName := "Updated Peer"
		newIP := netip.MustParseAddr("10.0.0.1")
		newMac, err := net.ParseMAC("11:22:33:44:55:66")
		if err != nil {
			t.Fatalf("parse mac: %v", err)
		}
		updated, err := q.UpdatePeerInfo(ctx, UpdatePeerInfoParams{
			WalletID:       wallet.ID,
			PeerWalletID:   peerWallet.ID,
			Name:           &newName,
			IpAddress:      &newIP,
			BtAddress:      newMac,
			ConnectionType: NullConnectionType{ConnectionType: ConnectionTypeBluetooth, Valid: true},
		})
		if err != nil {
			t.Fatalf("update peer info: %v", err)
		}
		if updated.Name == nil || *updated.Name != newName {
			t.Fatalf("expected updated name %q", newName)
		}

		trustedCount, err := q.CountTrustedPeers(ctx, wallet.ID)
		if err != nil {
			t.Fatalf("count trusted peers: %v", err)
		}
		if trustedCount != 1 {
			t.Fatalf("expected trusted peer count 1, got %d", trustedCount)
		}

		trustedPeers, err := q.ListTrustedPeers(ctx, wallet.ID)
		if err != nil {
			t.Fatalf("list trusted peers: %v", err)
		}
		if len(trustedPeers) != 1 {
			t.Fatalf("expected 1 trusted peer, got %d", len(trustedPeers))
		}
	})
}
