package common

import (
	"math"
	"strconv"

	"bjoernblessin.de/go-utils/util/assert"
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
// Environment variable is optional. If
//   - 0 is provided, 0 is returned.
//   - no value is provided, the default port is used.
//   - invalid value is provided, the application logs an error and exits.
func readAppPort() uint16 {
	portStr, found := env.ReadOptionalEnv(appPortEnvVar)

	if !found {
		return defaultAppPort
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		logger.Errorf("invalid PORT value: %s, must be between 0 and 65535", portStr)
	}

	assert.Assert(port < math.MaxUint16, "port value %d out of range", port)

	return uint16(port)
}

// readP2PPort is similar to readAppPort but for the P2P port.
func readP2PPort() uint16 {
	portStr, found := env.ReadOptionalEnv(p2pPortEnvVar)

	if !found {
		return defaultP2PPort
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		logger.Errorf("invalid P2P_PORT value: %s, must be between 0 and 65535", portStr)
	}

	assert.Assert(port < math.MaxUint16, "port value %d out of range", port)

	return uint16(port)
}

// SetAppPort sets the application port to the given value.
// Is needed for example for dynamic port assignment.
func SetAppPort(port uint16) {
	AppPort = port
}

func SetP2PPort(port uint16) {
	P2PPort = port
}
