package common

import "regexp"

// PrivateKeyWIFPattern Validation patterns as specified in OpenAPI and arc42 documentation.
// Base58Check format: starts with 5, followed by 50 Base58 characters (WIF format).
var PrivateKeyWIFPattern = regexp.MustCompile(`^5[1-9A-HJ-NP-Za-km-z]{50}$`)

// VsAddressPattern VSAddress validation: Base58Check encoded (starts with 1 for mainnet addresses).
var VsAddressPattern = regexp.MustCompile(`^1[1-9A-HJ-NP-Za-km-z]{25,34}$`)
