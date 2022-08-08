package libp2p

import (
	"fmt"
	mrand "math/rand"

	"github.com/libp2p/go-libp2p"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

// NewDummyHost creates an insecure libp2p host for test and demonstration purposes.
func NewDummyHost(id, basePort int) host.Host {
	sourceMultiAddr, priv := NewDummyHostID(id, basePort)

	h, err := libp2p.New(
		// Use the keypair we generated
		libp2p.Identity(priv),
		// Multiple listen addresses
		libp2p.DefaultTransports,
		libp2p.ListenAddrs(sourceMultiAddr),
	)
	if err != nil {
		panic(err)
	}

	return h
}

// NewDummyHostID generates a libp2p host address for test and demonstration purposes.
func NewDummyHostID(id, basePort int) (multiaddr.Multiaddr, libp2pcrypto.PrivKey) {
	rand := mrand.New(mrand.NewSource(int64(id))) // nolint

	priv, _, err := libp2pcrypto.GenerateKeyPairWithReader(libp2pcrypto.Ed25519, -1, rand)
	if err != nil {
		panic(err)
	}

	sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", basePort+id))
	if err != nil {
		panic(err)
	}

	return sourceMultiAddr, priv
}

// NewDummyMultiaddr generates a libp2p peer multiaddress for the host generated by NewDummyHost(id, basePort).
func NewDummyMultiaddr(id, basePort int) multiaddr.Multiaddr {
	sourceMultiAddr, priv := NewDummyHostID(id, basePort)

	peerID, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		panic(err)
	}

	peerInfo := peer.AddrInfo{
		ID:    peerID,
		Addrs: []multiaddr.Multiaddr{sourceMultiAddr},
	}
	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		panic(err)
	}

	if len(addrs) != 1 {
		panic(fmt.Errorf("wrong number of addresses %d", len(addrs)))
	}

	return addrs[0]
}