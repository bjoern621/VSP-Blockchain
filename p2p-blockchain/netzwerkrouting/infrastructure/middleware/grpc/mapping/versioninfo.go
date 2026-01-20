package mapping

import (
	"fmt"
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/handshake"

	"bjoernblessin.de/go-utils/util/assert"
)

// Mapping between proto ServiceType enum and domain ServiceType enum.
// The numeric values differ, so explicit mapping is required.
//
//	Proto Enum                      Value   Domain Enum                    Value
//	SERVICE_WALLET                  0       ServiceType_Netzwerkrouting    0
//	SERVICE_MINER                   1       ServiceType_BlockchainFull     1
//	SERVICE_BLOCKCHAIN_FULL         2       ServiceType_BlockchainSimple   2
//	SERVICE_NETZWERKROUTING         3       ServiceType_Miner              3

// serviceTypeFromProto converts a protobuf ServiceType to the domain ServiceType.
func serviceTypeFromProto(pbService pb.ServiceType) (common.ServiceType, bool) {
	switch pbService {
	case pb.ServiceType_SERVICE_NETZWERKROUTING:
		return common.ServiceType_Netzwerkrouting, true
	case pb.ServiceType_SERVICE_BLOCKCHAIN_FULL:
		return common.ServiceType_BlockchainFull, true
	case pb.ServiceType_SERVICE_WALLET:
		return common.ServiceType_Wallet, true
	case pb.ServiceType_SERVICE_MINER:
		return common.ServiceType_Miner, true
	default:
		return 0, false
	}
}

// serviceTypeToProto converts a domain ServiceType to the protobuf ServiceType.
// Fails if the ServiceType is unknown.
func serviceTypeToProto(service common.ServiceType) pb.ServiceType {
	switch service {
	case common.ServiceType_Netzwerkrouting:
		return pb.ServiceType_SERVICE_NETZWERKROUTING
	case common.ServiceType_BlockchainFull:
		return pb.ServiceType_SERVICE_BLOCKCHAIN_FULL
	case common.ServiceType_Wallet:
		return pb.ServiceType_SERVICE_WALLET
	case common.ServiceType_Miner:
		return pb.ServiceType_SERVICE_MINER
	default:
		assert.Never("unhandled ServiceType")
		return 0
	}
}

// VersionInfoFromProto converts protobuf VersionInfo to domain VersionInfo.
// Returns an error if the service types are invalid or fail validation.
func VersionInfoFromProto(info *pb.VersionInfo) (handshake.VersionInfo, netip.AddrPort, error) {
	var endpoint netip.AddrPort
	if info.ListeningEndpoint != nil {
		if ip, ok := netip.AddrFromSlice(info.ListeningEndpoint.IpAddress); ok {
			endpoint = netip.AddrPortFrom(ip, uint16(info.ListeningEndpoint.ListeningPort))
		}
	}

	services, err := serviceTypesFromProtoServiceTypes(info.SupportedServices)
	if err != nil {
		return handshake.VersionInfo{}, netip.AddrPort{}, fmt.Errorf("service type conversion failed: %w", err)
	}

	versionInfo := handshake.VersionInfo{
		Version: info.GetVersion(),
	}

	if err := versionInfo.TryAddService(services...); err != nil {
		return handshake.VersionInfo{}, netip.AddrPort{}, fmt.Errorf("version info construction failed: %w", err)
	}

	return versionInfo, endpoint, nil
}

func serviceTypesFromProtoServiceTypes(pbServices []pb.ServiceType) ([]common.ServiceType, error) {
	services := make([]common.ServiceType, 0, len(pbServices))
	for _, pbService := range pbServices {
		svc, ok := serviceTypeFromProto(pbService)
		if !ok {
			return nil, fmt.Errorf("unknown proto ServiceType: %v", pbService)
		}
		services = append(services, svc)
	}
	return services, nil
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
