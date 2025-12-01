package main

import (
	appcore "s3b/vsp-blockchain/p2p-blockchain/app/core"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/handshake"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc/peerregistry"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
)

func main() {
	logger.Infof("Running...")

	logger.Infof("Loglevel set to %v", logger.CurrentLevel())

	logger.Infof("Starting App server...")
	appServer := appcore.NewServer()
	err := appServer.Start(common.AppPort)
	if err != nil {
		logger.Warnf("couldn't start App server: %v", err)
	} else {
		addrPort, err := appServer.ListeningEndpoint()
		assert.IsNil(err)
		common.SetAppPort(addrPort.Port())
		logger.Infof("App server started on port %d", common.AppPort)
	}

	logger.Infof("Starting P2P server...")

	peerRegistry := peerregistry.NewPeerRegistry()
	handshakeInitiator := grpc.NewClient(peerRegistry)
	handshakeHandler := handshake.NewHandshakeService(handshakeInitiator)
	grpcServer := grpc.NewServer(handshakeHandler, peerRegistry)

	err = grpcServer.Start(common.P2PPort)
	if err != nil {
		logger.Warnf("couldn't start P2P server: %v", err)
	} else {
		addrPort, err := grpcServer.ListeningEndpoint()
		assert.IsNil(err)
		common.SetP2PPort(addrPort.Port())
		common.SetP2PListeningIpAddr(addrPort.Addr())
		logger.Infof("P2P server started on port %d", common.P2PPort)
	}

	select {}
}
