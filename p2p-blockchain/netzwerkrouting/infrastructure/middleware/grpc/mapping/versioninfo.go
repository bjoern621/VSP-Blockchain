package mapping

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/handshake"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
)

// Mapping between proto ServiceType enum and domain ServiceType enum.
// The numeric values differ, so explicit mapping is required.
//
//	Proto Enum                      Value   Domain Enum                    Value
//	SERVICE_WALLET                  0       ServiceType_Netzwerkrouting    0
//	SERVICE_MINER                   1       ServiceType_BlockchainFull     1
//	SERVICE_BLOCKCHAIN_FULL         2       ServiceType_BlockchainSimple   2
//	SERVICE_BLOCKCHAIN_SIMPLE       3       ServiceType_Wallet             3
//	SERVICE_NETZWERKROUTING         4       ServiceType_Miner              4

// serviceTypeFromProto converts a protobuf ServiceType to the domain ServiceType.
func serviceTypeFromProto(pbService pb.ServiceType) (peer.ServiceType, bool) {
	switch pbService {
	case pb.ServiceType_SERVICE_NETZWERKROUTING:
		return peer.ServiceType_Netzwerkrouting, true
	case pb.ServiceType_SERVICE_BLOCKCHAIN_FULL:
		return peer.ServiceType_BlockchainFull, true
	case pb.ServiceType_SERVICE_BLOCKCHAIN_SIMPLE:
		return peer.ServiceType_BlockchainSimple, true
	case pb.ServiceType_SERVICE_WALLET:
		return peer.ServiceType_Wallet, true
	case pb.ServiceType_SERVICE_MINER:
		return peer.ServiceType_Miner, true
	default:
		return 0, false
	}
}

// serviceTypeToProto converts a domain ServiceType to the protobuf ServiceType.
// Fails if the ServiceType is unknown.
func serviceTypeToProto(service peer.ServiceType) pb.ServiceType {
	switch service {
	case peer.ServiceType_Netzwerkrouting:
		return pb.ServiceType_SERVICE_NETZWERKROUTING
	case peer.ServiceType_BlockchainFull:
		return pb.ServiceType_SERVICE_BLOCKCHAIN_FULL
	case peer.ServiceType_BlockchainSimple:
		return pb.ServiceType_SERVICE_BLOCKCHAIN_SIMPLE
	case peer.ServiceType_Wallet:
		return pb.ServiceType_SERVICE_WALLET
	case peer.ServiceType_Miner:
		return pb.ServiceType_SERVICE_MINER
	default:
		assert.Never("unhandled ServiceType")
		return 0
	}
}

// VersionInfoFromProto converts protobuf VersionInfo to domain VersionInfo.
func VersionInfoFromProto(info *pb.VersionInfo) (handshake.VersionInfo, netip.AddrPort) {
	var endpoint netip.AddrPort
	if info.ListeningEndpoint != nil {
		if ip, ok := netip.AddrFromSlice(info.ListeningEndpoint.IpAddress); ok {
			endpoint = netip.AddrPortFrom(ip, uint16(info.ListeningEndpoint.ListeningPort))
		}
	}

	versionInfo := handshake.VersionInfo{
		Version: info.GetVersion(),
	}

	for _, pbService := range info.SupportedServices {
		if svc, ok := serviceTypeFromProto(pbService); ok {
			// todo verfiy order
			versionInfo.AddService(svc)
		} else {
			logger.Warnf("unknown proto ServiceType: %v", pbService)
		}
	}

	return versionInfo, endpoint
}

// VersionInfoToProto converts domain VersionInfo to protobuf VersionInfo.
func VersionInfoToProto(info handshake.VersionInfo, addrPort netip.AddrPort) *pb.VersionInfo {
	pbInfo := &pb.VersionInfo{
		Version: info.Version,
		ListeningEndpoint: &pb.Endpoint{
			IpAddress:     addrPort.Addr().AsSlice(),
			ListeningPort: uint32(addrPort.Port()),
		},
	}
	for _, service := range info.SupportedServices() {
		pbInfo.SupportedServices = append(pbInfo.SupportedServices, serviceTypeToProto(service))
	}
	return pbInfo
}
