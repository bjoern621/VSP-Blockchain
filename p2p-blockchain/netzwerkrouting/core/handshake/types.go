package handshake

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"

	"bjoernblessin.de/go-utils/util/assert"
)

type VersionInfo struct {
	Version           string
	SupportedServices []peer.ServiceType
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
	localSupportedService := []peer.ServiceType{}
	for _, svcString := range common.EnabledTeilsystemeNames() {
		svc, ok := peer.ParseServiceType(svcString)
		assert.Assert(ok, "enabled service is not a valid ServiceType:", svcString)
		localSupportedService = append(localSupportedService, svc)
	}

	return VersionInfo{
		Version:           common.VersionString,
		SupportedServices: localSupportedService,
	}
}
