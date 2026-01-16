package discovery

// DiscoveryService provides peer discovery functionality.
// This includes (1) querying a registry for peers and (2) asking neighbors for their known peers.
type DiscoveryService struct {
	querier          RegistryQuerier
	addrMsgSender    AddrMsgSender
	peerRetriever    peerRetriever
	getAddrMsgSender GetAddrMsgSender
}

// NewDiscoveryService creates a new DiscoveryService.
func NewDiscoveryService(
	querier RegistryQuerier,
	addrMsgSender AddrMsgSender,
	peerRetriever peerRetriever,
	getAddrMsgSender GetAddrMsgSender,
) *DiscoveryService {
	return &DiscoveryService{
		querier:          querier,
		addrMsgSender:    addrMsgSender,
		peerRetriever:    peerRetriever,
		getAddrMsgSender: getAddrMsgSender,
	}
}
