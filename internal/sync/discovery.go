package sync

import "fmt"

type Discovery struct {
	rendezvousTag string
	peerID        string
}

func NewDiscovery(rendezvousTag, peerID string) *Discovery {
	return &Discovery{
		rendezvousTag: rendezvousTag,
		peerID:        peerID,
	}
}

func (d *Discovery) Start() error {
	return fmt.Errorf("discovery not implemented: requires libp2p DHT integration")
}

func (d *Discovery) Stop() error {
	return nil
}
