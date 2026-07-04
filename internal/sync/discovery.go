package sync

import (
	"context"

	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
)

type Discovery struct {
	rendezvous  string
	host        host.Host
	dht         *dht.IpfsDHT
	routingDisc *routing.RoutingDiscovery
}

func NewDiscovery(h host.Host, d *dht.IpfsDHT, rendezvousTag string) *Discovery {
	return &Discovery{
		rendezvous:  rendezvousTag,
		host:        h,
		dht:         d,
		routingDisc: routing.NewRoutingDiscovery(d),
	}
}

func (d *Discovery) Advertise(ctx context.Context) {
	util.Advertise(ctx, d.routingDisc, d.rendezvous)
}

func (d *Discovery) FindPeers(ctx context.Context) ([]peer.AddrInfo, error) {
	return util.FindPeers(ctx, d.routingDisc, d.rendezvous)
}
