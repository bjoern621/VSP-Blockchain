package main

import (
	"math/rand"
	"sync"
	"time"

	"bjoernblessin.de/go-utils/util/logger"
)

// PeerState represents the current state of a peer.
type PeerState int

const (
	// StateNew indicates a peer that has been discovered but not yet connected.
	StateNew PeerState = iota
	// StateConnecting indicates a peer that is currently being connected to.
	StateConnecting
	// StateKnown indicates a peer that has been successfully connected to and verified.
	StateKnown
)

func (s PeerState) String() string {
	switch s {
	case StateNew:
		return "new"
	case StateConnecting:
		return "connecting"
	case StateKnown:
		return "known"
	default:
		return "unknown"
	}
}

// PeerInfo holds information about a discovered peer.
type PeerInfo struct {
	IP           string
	Port         int32
	State        PeerState
	LastSeen     time.Time
	DiscoveredAt time.Time
}

// PeerManager manages the lifecycle and state of discovered peers.
type PeerManager struct {
	mu       sync.RWMutex
	peers    map[string]*PeerInfo
	knownTTL time.Duration
}

// NewPeerManager creates a new PeerManager with the specified TTL for known peers.
func NewPeerManager(knownTTL time.Duration) *PeerManager {
	logger.Debugf("creating peer manager with TTL=%s", knownTTL)
	return &PeerManager{
		peers:    make(map[string]*PeerInfo),
		knownTTL: knownTTL,
	}
}

// AddPeer adds a new peer if it does not exist or is expired.
// Returns true if the peer was added as new.
func (pm *PeerManager) AddPeer(ip string, port int32) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	now := time.Now()

	if existing, ok := pm.peers[ip]; ok {
		if existing.State == StateKnown && now.Sub(existing.LastSeen) < pm.knownTTL {
			return false
		}
		if existing.State == StateConnecting {
			return false
		}
		if existing.State == StateNew {
			return false
		}
		existing.State = StateNew
		existing.DiscoveredAt = now
		return true
	}

	pm.peers[ip] = &PeerInfo{
		IP:           ip,
		Port:         port,
		State:        StateNew,
		DiscoveredAt: now,
	}
	logger.Debugf("added new peer %s:%d", ip, port)
	return true
}

// AddPeers adds multiple peers. Returns the count of newly added peers.
func (pm *PeerManager) AddPeers(ips map[string]struct{}, port int32) int {
	added := 0
	for ip := range ips {
		if pm.AddPeer(ip, port) {
			added++
		}
	}
	return added
}

// GetNextNewPeer returns a peer in StateNew and transitions it to StateConnecting.
// Returns nil if no StateNew peers are available.
func (pm *PeerManager) GetNextNewPeer() *PeerInfo {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, peer := range pm.peers {
		if peer.State == StateNew {
			peer.State = StateConnecting
			return &PeerInfo{
				IP:           peer.IP,
				Port:         peer.Port,
				State:        StateConnecting,
				DiscoveredAt: peer.DiscoveredAt,
			}
		}
	}
	logger.Tracef("no new peers available for connection")
	return nil
}

// GetExpiredKnownPeer returns a known peer whose TTL has expired for re-verification.
// Returns nil if no expired known peers are available.
func (pm *PeerManager) GetExpiredKnownPeer() *PeerInfo {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	now := time.Now()
	for _, peer := range pm.peers {
		if peer.State == StateKnown && now.Sub(peer.LastSeen) >= pm.knownTTL {
			peer.State = StateConnecting
			logger.Debugf("peer %s TTL expired (last seen %s ago), re-verifying", peer.IP, now.Sub(peer.LastSeen))
			return &PeerInfo{
				IP:       peer.IP,
				Port:     peer.Port,
				State:    StateConnecting,
				LastSeen: peer.LastSeen,
			}
		}
	}
	return nil
}

// MarkConnected marks a peer as successfully connected and updates LastSeen.
func (pm *PeerManager) MarkConnected(ip string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if peer, ok := pm.peers[ip]; ok {
		peer.State = StateKnown
		peer.LastSeen = time.Now()
	}
}

// MarkFailed removes a peer that failed to connect.
func (pm *PeerManager) MarkFailed(ip string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	logger.Debugf("peer %s failed verification, removing", ip)
	delete(pm.peers, ip)
}

// ResetToNew resets a connecting peer back to new state (for retries).
func (pm *PeerManager) ResetToNew(ip string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if peer, ok := pm.peers[ip]; ok && peer.State == StateConnecting {
		peer.State = StateNew
	}
}

// GetKnownPeers returns all peers in StateKnown that are within the TTL.
func (pm *PeerManager) GetKnownPeers() []*PeerInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	now := time.Now()
	result := make([]*PeerInfo, 0)
	for _, peer := range pm.peers {
		if peer.State == StateKnown && now.Sub(peer.LastSeen) < pm.knownTTL {
			result = append(result, &PeerInfo{
				IP:       peer.IP,
				Port:     peer.Port,
				State:    peer.State,
				LastSeen: peer.LastSeen,
			})
		}
	}
	return result
}

// GetRandomKnownPeerIPsFilteredByPort returns IP addresses of a random subset of known peers
// that use the given port.
func (pm *PeerManager) GetRandomKnownPeerIPsFilteredByPort(port int32, count int) map[string]struct{} {
	known := pm.GetKnownPeers()
	filtered := make([]*PeerInfo, 0, len(known))
	for _, p := range known {
		if p.Port == port {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) > count {
		rand.Shuffle(len(filtered), func(i, j int) {
			filtered[i], filtered[j] = filtered[j], filtered[i]
		})
		filtered = filtered[:count]
	}

	result := make(map[string]struct{}, len(filtered))
	for _, p := range filtered {
		result[p.IP] = struct{}{}
	}
	return result
}

// CleanupExpired removes peers that have been expired for longer than the TTL.
func (pm *PeerManager) CleanupExpired() int {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	now := time.Now()
	removed := 0
	for ip, peer := range pm.peers {
		if peer.State == StateKnown && now.Sub(peer.LastSeen) >= 2*pm.knownTTL {
			logger.Debugf("cleaning up expired peer %s (last seen %s ago)", ip, now.Sub(peer.LastSeen))
			delete(pm.peers, ip)
			removed++
		}
	}
	if removed > 0 {
		logger.Debugf("cleaned up %d expired peers", removed)
	}
	return removed
}

// Stats returns statistics about the peer manager state.
func (pm *PeerManager) Stats() (total, newCount, connecting, known int) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, peer := range pm.peers {
		total++
		switch peer.State {
		case StateNew:
			newCount++
		case StateConnecting:
			connecting++
		case StateKnown:
			known++
		}
	}
	return
}
