package common

import "bjoernblessin.de/go-utils/util/assert"

type Direction int

const (
	DirectionInbound Direction = iota
	DirectionOutbound
	DirectionBoth
)

func (d Direction) String() string {
	switch d {
	case DirectionInbound:
		return "inbound"
	case DirectionOutbound:
		return "outbound"
	case DirectionBoth:
		return "both"
	default:
		assert.Never("unhandled Direction")
		return "unknown"
	}
}
