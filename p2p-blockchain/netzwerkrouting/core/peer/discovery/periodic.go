package discovery

import (
	"math/rand"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"time"

	"bjoernblessin.de/go-utils/util/logger"
)

const (
	// DefaultGossipDiscoveryInterval is how often to send getaddr to connected peers.
	DefaultGossipDiscoveryInterval = 5 * time.Minute
	// DefaultGossipDiscoveryPeers is the number of peers to query during gossip discovery.
	DefaultGossipDiscoveryPeers = 3
)

type registryQuerier interface {
	// GetPeers queries the registry for available peer IDs.
	GetPeers()
}

// PeriodicDiscoveryService manages periodic peer discovery.
// This includes gossip as well as querying the registry.
// It periodically sends getaddr messages to random connected peers to discover new peers.
type PeriodicDiscoveryService struct {
	peerRetriever    peerRetriever
	getAddrMsgSender GetAddrMsgSender

	registryQuerier registryQuerier

	discoveryInterval time.Duration
	discoveryPeers    int

	stopChan      chan struct{}
	ticker        *time.Ticker
	lastDiscovery time.Time
}

// NewPeriodicDiscoveryService creates a new GossipDiscoveryService.
func NewPeriodicDiscoveryService(
	peerRetriever peerRetriever,
	getAddrMsgSender GetAddrMsgSender,
	registryQuerier registryQuerier,
) *PeriodicDiscoveryService {
	return &PeriodicDiscoveryService{
		peerRetriever:     peerRetriever,
		getAddrMsgSender:  getAddrMsgSender,
		registryQuerier:   registryQuerier,
		discoveryInterval: DefaultGossipDiscoveryInterval,
		discoveryPeers:    DefaultGossipDiscoveryPeers,
		stopChan:          make(chan struct{}),
	}
}

// Start begins the periodic gossip discovery.
func (s *PeriodicDiscoveryService) Start() {
	s.ticker = time.NewTicker(s.discoveryInterval)
	go s.run()
}

// Stop halts the periodic gossip discovery.
func (s *PeriodicDiscoveryService) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
		select {
		case <-s.stopChan:
			// Channel already closed, do nothing
		default:
			// Channel not yet closed, close it
			close(s.stopChan)
		}
	}
}

// run periodically performs gossip discovery.
func (s *PeriodicDiscoveryService) run() {
	for {
		select {
		case <-s.ticker.C:
			s.performGossipDiscovery()
			s.performRegistryDiscovery()
		case <-s.stopChan:
			return
		}
	}
}

// performRegistryDiscovery queries the registry for new peers.
func (s *PeriodicDiscoveryService) performRegistryDiscovery() {
	s.registryQuerier.GetPeers()
}

// performGossipDiscovery sends getaddr to random connected peers to discover new peers.
func (s *PeriodicDiscoveryService) performGossipDiscovery() {
	if s.lastDiscovery.IsZero() {
		// First run, skip to allow time for initial connections
		s.lastDiscovery = time.Now()
		return
	}

	connectedPeers := s.peerRetriever.GetAllConnectedPeers()

	if len(connectedPeers) == 0 {
		logger.Debugf("No connected peers for gossip discovery")
		return
	}

	// Select random peers to query for gossip discovery
	numToQuery := min(len(connectedPeers), s.discoveryPeers)

	// Shuffle and select random peers
	shuffled := make([]common.PeerId, len(connectedPeers))
	copy(shuffled, connectedPeers)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	selectedPeers := shuffled[:numToQuery]

	// Send getaddr to selected peers
	for _, peerID := range selectedPeers {
		s.getAddrMsgSender.SendGetAddr(peerID)
		logger.Infof("Sent getaddr to peer %s for gossip discovery", peerID)
	}

	s.lastDiscovery = time.Now()
}
