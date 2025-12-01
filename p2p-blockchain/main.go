package main

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/app/core"
	"s3b/vsp-blockchain/p2p-blockchain/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/config"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"

	"bjoernblessin.de/go-utils/util/logger"
)

func main() {
	logger.Infof("Running...")

	appPort := common.AppPort
	p2pPort := common.P2PPort

	config.Init(p2pPort, netip.MustParseAddr("::1"))

	logger.Infof("Starting App server on port %d...", appPort)
	err := core.NewServer().Start(appPort)
	if err != nil {
		logger.Warnf("couldn't start App server: %v", err)
	}

	logger.Infof("Starting P2P server on port %d...", p2pPort)
	err = api.NewServer().Start(p2pPort)
	if err != nil {
		logger.Warnf("couldn't start P2P server: %v", err)
	}

	select {}
}
