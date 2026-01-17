package common

import "bjoernblessin.de/go-utils/util/assert"

type Direction int

const (
	DirectionUnknown Direction = iota
	DirectionInbound
	DirectionOutbound
	DirectionBoth
)

func (d Direction) String() string {
	switch d {
	case DirectionUnknown:
		return "unknown"
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
