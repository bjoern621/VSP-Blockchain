package common

import (
	"math"
	"math/rand/v2"
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
//   - 0 is provided, a random port between 49152 and 65535 is used.
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

	if port == 0 {
		port = uint64(rand.IntN(math.MaxUint16-49152+1) + 49152) // [49152, 65535]
	}

	assert.Assert(port >= 1 && port < math.MaxUint16, "port value %d out of range", port)

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

	if port == 0 {
		port = uint64(rand.IntN(math.MaxUint16-49152+1) + 49152) // [49152, 65535]
	}

	assert.Assert(port >= 1 && port < math.MaxUint16, "port value %d out of range", port)

	return uint16(port)
}
