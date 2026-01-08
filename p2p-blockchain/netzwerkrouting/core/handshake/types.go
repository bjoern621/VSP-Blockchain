package handshake

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"

	"bjoernblessin.de/go-utils/util/assert"
)

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

	var services []common.ServiceType
	for _, svcString := range common.EnabledTeilsystemeNames() {
		svc, ok := common.ParseServiceType(svcString)
		assert.Assert(ok, "enabled service is not a valid ServiceType:", svcString)
		services = append(services, svc)
	}
	info.AddService(services...)

	return info
}
