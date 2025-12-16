package updater

import (
	"math/rand"
	"net/netip"
	"sort"
	"time"
)

// generateRandomIPv4s generates random IPv4 addresses within the given CIDR prefix.
// Used for debug/testing purposes to simulate many seed nodes.
func generateRandomIPv4s(prefix netip.Prefix, count int) []string {
	addr := prefix.Masked().Addr()
	base := addr.As4()
	bits := 32 - prefix.Bits()
	max := uint32(1) << uint32(bits)
	if max <= 2 {
		return []string{addr.String()}
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	seen := map[uint32]struct{}{}
	res := make([]string, 0, count)

	for len(res) < count {
		offset := uint32(rng.Intn(int(max-2))) + 1
		if _, ok := seen[offset]; ok {
			continue
		}
		seen[offset] = struct{}{}

		ip := netip.AddrFrom4([4]byte{base[0], base[1], base[2], base[3]}).Next()
		for i := uint32(1); i < offset; i++ {
			ip = ip.Next()
		}
		res = append(res, ip.String())
	}

	sort.Strings(res)
	return res
}
