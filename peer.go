package main

import "sync"

// PeerManager ...
type PeerManager struct {
	Peers map[string]string
	Num   int
	mtx   *sync.Mutex
}

// PeerManagerInfo ...
type PeerInfo struct {
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
			go pm.addPeer(peers)
		case <-SToPGetPI:
			go pm.returnServerPeerInfo()
		case <-CToPGetPI:
			go pm.returnClientPeerInfo()
		default:
		}
	}
}

// returnClientPeerInfo return PeerInfo to Client
func (pm *PeerManager) returnClientPeerInfo() {
	pi := pm.getPeerInfo()
	PToCPI <- pi
}

// returnServerPeerInfo return PeerInfo to Server
func (pm *PeerManager) returnServerPeerInfo() {
	pi := pm.getPeerInfo()
	PToSPI <- pi
}

// AddPeer ...
func (pm *PeerManager) addPeer(peers []string) {
	pm.mtx.Lock()
	defer pm.mtx.Unlock()

	for _, v := range peers {
		if pm.Peers[v] == "" {
			pm.Peers[v] = v
			pm.Num++
		}
	}
}

// DeletePeer ...
func (pm *PeerManager) deletePeer() {
	pm.mtx.Lock()
	defer pm.mtx.Unlock()

}

// GetPeerManagerInfo ...
func (pm *PeerManager) getPeerInfo() *PeerInfo {
	pm.mtx.Lock()
	defer pm.mtx.Unlock()

	return &PeerInfo{pm.Peers, pm.Num}
}
