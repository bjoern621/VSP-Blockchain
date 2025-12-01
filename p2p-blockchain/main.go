package main

import (
	appcore "s3b/vsp-blockchain/p2p-blockchain/app/core"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	ncore "s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core"
	ninfrastructure "s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
)

func main() {
	logger.Infof("Running...")

	appPort := common.AppPort
	p2pPort := common.P2PPort

	logger.Infof("Starting App server...")
	appServer := appcore.NewServer()
	err := appServer.Start(appPort)
	if err != nil {
		logger.Warnf("couldn't start App server: %v", err)
	} else {
		logger.Infof("App server started on port %d", appPort)
		addrPort, err := appServer.ListeningEndpoint()
		assert.IsNil(err)
		common.SetAppPort(addrPort.Port())
	}

	logger.Infof("Starting P2P server...", p2pPort)
	handshakeSerivce := ncore.NewHandshakeService()
	grpcServer := ninfrastructure.NewServer(handshakeSerivce)
	err = grpcServer.Start(p2pPort)
	if err != nil {
		logger.Warnf("couldn't start P2P server: %v", err)
	} else {
		logger.Infof("P2P server started on port %d", p2pPort)
		addrPort, err := grpcServer.ListeningEndpoint()
		assert.IsNil(err)
		common.SetP2PPort(addrPort.Port())
		common.SetP2PListeningIpAddr(addrPort.Addr())
	}

	select {}
}
