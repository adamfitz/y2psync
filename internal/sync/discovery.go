package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
)

type Discovery struct {
	rendezvous string
	host       host.Host
	dht        *dht.IpfsDHT
	discovery  *routing.RoutingDiscovery

	peerCh chan peer.AddrInfo
	cancel context.CancelFunc
	done   chan struct{}
}

func NewDiscovery(h host.Host, d *dht.IpfsDHT, rendezvousTag string) *Discovery {
	rd := routing.NewRoutingDiscovery(d)

	return &Discovery{
		rendezvous: rendezvousTag,
		host:       h,
		dht:        d,
		discovery:  rd,
		peerCh:     make(chan peer.AddrInfo, 64),
		done:       make(chan struct{}),
	}
}

func (d *Discovery) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	d.cancel = cancel

	util.Advertise(ctx, d.discovery, d.rendezvous)

	peerList, err := util.FindPeers(ctx, d.discovery, d.rendezvous)
	if err != nil {
		cancel()
		return fmt.Errorf("find peers: %w", err)
	}

	go d.discoverLoop(ctx, peerList)

	return nil
}

func (d *Discovery) discoverLoop(ctx context.Context, peerList []peer.AddrInfo) {
	defer close(d.done)

	for _, pi := range peerList {
		if pi.ID == d.host.ID() {
			continue
		}
		if len(pi.Addrs) == 0 {
			continue
		}
		select {
		case d.peerCh <- pi:
		case <-ctx.Done():
			return
		}
	}

	<-ctx.Done()
}

func (d *Discovery) Peers() <-chan peer.AddrInfo {
	return d.peerCh
}

func (d *Discovery) Stop() {
	if d.cancel != nil {
		d.cancel()
	}
	select {
	case <-d.done:
	case <-time.After(2 * time.Second):
	}
}
