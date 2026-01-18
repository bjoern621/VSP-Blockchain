package common

// Error type constants for reject messages.
// These correspond to the ErrorType enum in the protobuf definition (proto/netzwerkrouting.proto).
// Core layer uses constants instead of protobuf enums to maintain clean architecture.

const (
	// ErrorTypeRejectMalformed indicates a message that cannot be parsed or has an incorrect format
	ErrorTypeRejectMalformed = 0

	// ErrorTypeRejectInvalid indicates a well-formed message that fails validation logic
	ErrorTypeRejectInvalid = 1

	// ErrorTypeRejectHolddown indicates a peer is being temporarily rejected due to repeated bad behavior
	ErrorTypeRejectHolddown = 2

	// ErrorTypeRejectNotConnected indicates that a message was received from a peer that is not in an established connection state
	ErrorTypeRejectNotConnected = 3
)
