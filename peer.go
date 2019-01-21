package main

import "sync"

// PeerManager ...
type PeerManager struct {
	Peers map[string]string
	Num   int
	mtx   *sync.Mutex
}

// PeerManagerInfo ...
type PeerManagerInfo struct {
	Peers map[string]string
	Num   int
}

// NewPeerManager returns a new PeerManager
func NewPeerManager() *PeerManager {
	return &PeerManager{make(map[string]string), 0, new(sync.Mutex)}
}

// Processor ...
func (pm *PeerManager) Processor() {
	for {
		select {
		case peers := <-SToPPeer:
			go pm.AddPeer(peers)
		case <-SToPGetPM:
			go pm.GetPeerManagerInfo()
		default:
		}
	}
}

// AddPeer ...
func (pm *PeerManager) AddPeer(peers []string) {
	pm.mtx.Lock()
	defer pm.mtx.Unlock()

	for _, v := range peers {
		if pm.Peers[v] == "" {
			pm.Peers[v] = v
		}
	}
}

// DeletePeer ...
func (pm *PeerManager) DeletePeer() {

}

// GetPeerManagerInfo ...
func (pm *PeerManager) GetPeerManagerInfo() {
	pm.mtx.Lock()
	defer pm.mtx.Unlock()

	npm := &PeerManagerInfo{pm.Peers, pm.Num}
	PToSSendPM <- npm
}