package main

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/app/core"
	"s3b/vsp-blockchain/p2p-blockchain/internal/config"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
	"strconv"

	"bjoernblessin.de/go-utils/util/env"
	"bjoernblessin.de/go-utils/util/logger"
)

const defaultAppPort = 50050
const defaultP2PPort = 50051
const portEnvVar = "APP_PORT"
const p2pPortEnvVar = "P2P_PORT"

func getPortFromEnv() uint16 {
	portStr := env.ReadNonEmptyRequiredEnv(portEnvVar)

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		logger.Errorf("invalid PORT value: %s, must be between 0 and 65535", portStr)
	}

	if port == 0 {
		port = defaultAppPort
	}

	return uint16(port)
}

func getP2PPortFromEnv() uint16 {
	portStr := env.ReadNonEmptyRequiredEnv(p2pPortEnvVar)

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		logger.Errorf("invalid P2P_PORT value: %s, must be between 0 and 65535", portStr)
	}

	if port == 0 {
		port = defaultP2PPort
	}

	return uint16(port)
}

func main() {
	logger.Infof("Running...")

	port := getPortFromEnv()
	p2pPort := getP2PPortFromEnv()

	config.Init(p2pPort, netip.MustParseAddr("::1"))

	logger.Infof("Starting App server on port %d...", port)
	err := core.NewServer().Start(port)
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
