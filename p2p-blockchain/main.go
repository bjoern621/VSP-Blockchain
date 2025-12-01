package main

import (
	"net/netip"
	appcore "s3b/vsp-blockchain/p2p-blockchain/app/core"
	"s3b/vsp-blockchain/p2p-blockchain/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/config"
	ncore "s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core"
	ninfrastructure "s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc"

	"bjoernblessin.de/go-utils/util/logger"
)

func main() {
	logger.Infof("Running...")

	appPort := common.AppPort
	p2pPort := common.P2PPort

	config.Init(p2pPort, netip.MustParseAddr("::1"))

	logger.Infof("Starting App server on port %d...", appPort)
	err := appcore.NewServer().Start(appPort)
	if err != nil {
		logger.Warnf("couldn't start App server: %v", err)
	}

	logger.Infof("Starting P2P server on port %d...", p2pPort)
	handshakeSerivce := ncore.NewHandshakeService()
	grpcServer := ninfrastructure.NewServer(handshakeSerivce)
	err = grpcServer.Start(p2pPort)
	if err != nil {
		logger.Warnf("couldn't start P2P server: %v", err)
	}

	select {}
}
