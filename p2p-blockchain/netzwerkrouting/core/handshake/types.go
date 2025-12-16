package handshake

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
)

type VersionInfo struct {
	Version string
	// supportedServices holds the list of services supported by the peer.
	// It is guaranteed to follow all domain rules.
	supportedServices []peer.ServiceType
}

// SupportedServices returns a copy of the supported services slice.
func (v *VersionInfo) SupportedServices() []peer.ServiceType {
	return append([]peer.ServiceType(nil), v.supportedServices...)
}

// AddService adds a service to supportedServices with domain rule validation.
// Rules:
//   - No duplicate services allowed.
//   - blockchain_full and blockchain_simple are mutually exclusive.
//   - wallet requires blockchain_full or blockchain_simple.
//   - miner requires blockchain_full or blockchain_simple.
func (v *VersionInfo) AddService(svc ...peer.ServiceType) {
	for _, s := range svc {
		v.addService(s)
	}
}

func (v *VersionInfo) addService(svc peer.ServiceType) {
	for _, existing := range v.supportedServices {
		assert.Assert(existing != svc, "duplicate service:", svc)
	}

	if svc == peer.ServiceType_BlockchainFull {
		for _, existing := range v.supportedServices {
			assert.Assert(existing != peer.ServiceType_BlockchainSimple,
				"blockchain_full and blockchain_simple are mutually exclusive")
		}
	}

	if svc == peer.ServiceType_BlockchainSimple {
		for _, existing := range v.supportedServices {
			assert.Assert(existing != peer.ServiceType_BlockchainFull,
				"blockchain_full and blockchain_simple are mutually exclusive")
		}
	}

	v.supportedServices = append(v.supportedServices, svc)
}

// ValidateServices checks that the current SupportedServices satisfy all domain rules.
// Call this after all services have been added.
func (v *VersionInfo) ValidateServices() {
	hasBlockchain := false
	for _, svc := range v.supportedServices {
		if svc == peer.ServiceType_BlockchainFull || svc == peer.ServiceType_BlockchainSimple {
			hasBlockchain = true
			break
		}
	}

	for _, svc := range v.supportedServices {
		if svc == peer.ServiceType_Wallet {
			assert.Assert(hasBlockchain, "wallet requires blockchain_full or blockchain_simple")
		}
		if svc == peer.ServiceType_Miner {
			assert.Assert(hasBlockchain, "miner requires blockchain_full or blockchain_simple")
		}
	}
}

// handshakeService implements HandshakeMsgHandler (for infrastructure) and HandshakeInitiator (for api) with the actual domain logic.
type handshakeService struct {
	handshakeMsgSender HandshakeMsgSender
	peerStore          *peer.PeerStore
}

func NewHandshakeService(handshakeMsgSender HandshakeMsgSender, peerStore *peer.PeerStore) *handshakeService {
	return &handshakeService{
		handshakeMsgSender: handshakeMsgSender,
		peerStore:          peerStore,
	}
}

// NewLocalVersionInfo creates a VersionInfo struct with the local node's version and supported services.
func NewLocalVersionInfo() VersionInfo {
	info := VersionInfo{
		Version: common.VersionString,
	}

	logger.Warnf("systems %v", common.EnabledTeilsystemeNames())

	for _, svcString := range common.EnabledTeilsystemeNames() {
		svc, ok := peer.ParseServiceType(svcString)
		logger.Warnf("adding service %v", svc)
		assert.Assert(ok, "enabled service is not a valid ServiceType:", svcString)
		info.AddService(svc)
	}

	info.ValidateServices()

	return info
}
