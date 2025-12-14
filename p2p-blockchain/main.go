package main

import (
	appcore "s3b/vsp-blockchain/p2p-blockchain/app/core"
	appgrpc "s3b/vsp-blockchain/p2p-blockchain/app/infrastructure/grpc"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/handshake"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc/networkinfo"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/registry"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
)

func main() {
	logger.Infof("Running...")

	logger.Infof("Loglevel set to %v", logger.CurrentLevel())

	peerStore := peer.NewPeerStore()
	networkInfoRegistry := networkinfo.NewNetworkInfoRegistry(peerStore)
	grpcClient := grpc.NewClient(networkInfoRegistry)
	handshakeService := handshake.NewHandshakeService(grpcClient, peerStore)
	handshakeAPI := api.NewHandshakeAPIService(networkInfoRegistry, peerStore, handshakeService)
	networkRegistryAPI := api.NewNetworkRegistryService(networkInfoRegistry, peerStore)
	registryQuerier := registry.NewDNSRegistryQuerier(networkInfoRegistry)
	queryRegistryAPI := api.NewQueryRegistryService(registryQuerier)

	if common.AppEnabled() {
		logger.Infof("Starting App server...")

		connService := appcore.NewConnectionEstablishmentService(handshakeAPI)
		internalViewService := appcore.NewInternsalViewService(networkRegistryAPI)
		queryRegistryService := appcore.NewQueryRegistryService(queryRegistryAPI)
		appServer := appgrpc.NewServer(connService, internalViewService, queryRegistryService)

		err := appServer.Start(common.AppPort())
		if err != nil {
			logger.Warnf("couldn't start App server: %v", err)
		} else {
			addrPort, err := appServer.ListeningEndpoint()
			assert.IsNil(err)
			common.SetAppPort(addrPort.Port())
			logger.Infof("App server started on port %v", addrPort)
		}
	}

	logger.Infof("Starting P2P server...")

	grpcServer := grpc.NewServer(handshakeService, networkInfoRegistry)

	err := grpcServer.Start(common.P2PPort())
	if err != nil {
		logger.Warnf("couldn't start P2P server: %v", err)
	} else {
		addrPort, err := grpcServer.ListeningEndpoint()
		assert.IsNil(err)
		common.SetP2PPort(addrPort.Port())
		common.SetP2PListeningIpAddr(addrPort.Addr())
		logger.Infof("P2P server started on port %v", addrPort)
	}

	select {}
}
