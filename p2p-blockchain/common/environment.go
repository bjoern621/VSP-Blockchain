package common

import (
	"strconv"

	"bjoernblessin.de/go-utils/util/env"
	"bjoernblessin.de/go-utils/util/logger"
)

const (
	appPortEnvVar = "APP_PORT"
	p2pPortEnvVar = "P2P_PORT"
)

var (
	AppPort uint16
	P2PPort uint16
)

// init reads all required environment variables at startup.
// Read values are stored in package-level variables for easy access.
func init() {
	AppPort = readAppPort()
	P2PPort = readP2PPort()
}

// readAppPort reads the application port used by the app endpoint  from the environment variable APP_PORT.
// Environment variable is required. If 0 is provided, the default port is used.
func readAppPort() uint16 {
	portStr := env.ReadNonEmptyRequiredEnv(appPortEnvVar)

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		logger.Errorf("invalid PORT value: %s, must be between 0 and 65535", portStr)
	}

	if port == 0 {
		port = defaultAppPort
	}

	return uint16(port)
}

// readP2PPort is similar to readAppPort but for the P2P port.
func readP2PPort() uint16 {
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
