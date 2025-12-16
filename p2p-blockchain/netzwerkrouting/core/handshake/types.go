package handshake

import (
	"errors"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
	"slices"

	"bjoernblessin.de/go-utils/util/assert"
)

var (
	ErrDuplicateService            = errors.New("duplicate service")
	ErrMutuallyExclusiveBlockchain = errors.New("blockchain_full and blockchain_simple are mutually exclusive")
	ErrWalletRequiresBlockchain    = errors.New("wallet requires blockchain_full or blockchain_simple")
	ErrMinerRequiresBlockchain     = errors.New("miner requires blockchain_full or blockchain_simple")
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

// validateRequiresBlockchain checks that all domain rules are satisfied.
// Rules:
//   - wallet requires blockchain_full or blockchain_simple.
//   - miner requires blockchain_full or blockchain_simple.
func (v *VersionInfo) validateRequiresBlockchain() error {
	hasBlockchain := slices.Contains(v.supportedServices, peer.ServiceType_BlockchainFull) ||
		slices.Contains(v.supportedServices, peer.ServiceType_BlockchainSimple)

	if slices.Contains(v.supportedServices, peer.ServiceType_Wallet) && !hasBlockchain {
		return ErrWalletRequiresBlockchain
	}

	if slices.Contains(v.supportedServices, peer.ServiceType_Miner) && !hasBlockchain {
		return ErrMinerRequiresBlockchain
	}

	return nil
}

// TryAddService tries to add a service to supportedServices with domain rule validation.
// Returns true if the service was added successfully, false otherwise.
func (v *VersionInfo) TryAddService(svc ...peer.ServiceType) error {
	for _, s := range svc {
		err := v.addService(s)
		if err != nil {
			return err
		}
	}

	if err := v.validateRequiresBlockchain(); err != nil {
		return err
	}
	return nil
}

// AddService adds a service to supportedServices with domain rule validation.
// Panics on validation errors.
// Rules:
//   - No duplicate services allowed.
//   - blockchain_full and blockchain_simple are mutually exclusive.
//   - wallet requires blockchain_full or blockchain_simple.
//   - miner requires blockchain_full or blockchain_simple.
func (v *VersionInfo) AddService(svc ...peer.ServiceType) {
	for _, s := range svc {
		err := v.addService(s)
		assert.Assert(err == nil, err)
	}
	err := v.validateRequiresBlockchain()
	assert.Assert(err == nil, err)
}

func (v *VersionInfo) addService(svc peer.ServiceType) error {
	if slices.Contains(v.supportedServices, svc) {
		return ErrDuplicateService
	}

	if svc == peer.ServiceType_BlockchainFull {
		if slices.Contains(v.supportedServices, peer.ServiceType_BlockchainSimple) {
			return ErrMutuallyExclusiveBlockchain
		}
	}

	if svc == peer.ServiceType_BlockchainSimple {
		if slices.Contains(v.supportedServices, peer.ServiceType_BlockchainFull) {
			return ErrMutuallyExclusiveBlockchain
		}
	}

	v.supportedServices = append(v.supportedServices, svc)
	return nil
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

	for _, svcString := range common.EnabledTeilsystemeNames() {
		svc, ok := peer.ParseServiceType(svcString)
		assert.Assert(ok, "enabled service is not a valid ServiceType:", svcString)
		info.AddService(svc)
	}

	return info
}
