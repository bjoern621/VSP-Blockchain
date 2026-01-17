package common

type PeerId string

func (p PeerId) String() string {
	if p == "" {
		return "local miner"
	}
	return string(p)
}
