package src

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	tls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
)

type p2pHost struct {
	ctx       context.Context
	Host      host.Host
	KaDHT     *dht.IpfsDHT
	Discovery discovery.RoutingDiscovery
}

func EstablishP2P() *p2pHost {
	ctx := context.Background()
	nodehost, kdht := creatNode(ctx)
	bootsrapDHT(ctx, nodehost, kdht)
	routingDiscovery := discovery.NewRoutingDiscovery(kdht)
	return &p2pHost{
		ctx:       ctx,
		Host:      nodehost,
		KaDHT:     kdht,
		Discovery: routingDiscovery,
	}
}

func creatNode(ctx context.Context) (host.Host, *dht.IpfsDHT) {
	prvkey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		panic(err)
	}
	identity := libp2p.Identity(prvkey)
	tlstransport, err := tls.New(prvkey)
	if err != nil {
		panic(err)
	}
	security := libp2p.Security(tls.ID, tlstransport)
	transport := libp2p.Transport(tcp.NewTCPTransport)
	muladdr, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/0")
	listen := libp2p.ListenAddrs(muladdr)
	if err != nil {
		panic(err)
	}
	muxer := libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport)
	conn := libp2p.ConnectionManager(connmgr.NewConnManager(100, 400, time.Minute))
	nat := libp2p.NATPortMap()
	relay := libp2p.EnableAutoRelay()
	var kdht *dht.IpfsDHT
	routing := libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		kdht = setupKadDHT(ctx, h)
		return kdht, err
	})
	opts := libp2p.ChainOptions(identity, listen, security, transport, muxer, conn, nat, routing, relay)
	libhost, err := libp2p.New(ctx, opts)
	if err != nil {
		panic(err)
	}
	return libhost, kdht

}

func setupKadDHT(ctx context.Context, nodehost host.Host) *dht.IpfsDHT {
	dhtmode := dht.Mode(dht.ModeServer)
	bootsrappeers := dht.GetDefaultBootstrapPeerAddrInfos()
	dhtpeers := dht.BootstrapPeers(bootsrappeers...)

	kdht, err := dht.New(ctx, nodehost, dhtmode, dhtpeers)
	if err != nil {
		panic(err)
	}
	return kdht

}

func bootsrapDHT(ctx context.Context, nodehost host.Host, kdht *dht.IpfsDHT) {
	if err := kdht.Bootstarp(ctx); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	var connectedBootpeers int
	var totalbootpeers int
	for _, peeraddr := range dht.DefaultBootstrapPeers {
		peerInfor, _ := peer.AddrInfoFromP2pAddr(peeraddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := nodehost.Connect(ctx, *peerInfor); err != nil {
				totalbootpeers++
			} else {
				connectedBootpeers++
				totalbootpeers++
			}
		}()
	}
	wg.Wait()
	fmt.Println("Connected to ", connectedBootpeers, "of ", totalbootpeers)

}
