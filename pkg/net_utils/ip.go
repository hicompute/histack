package netutils

import (
	"fmt"
	"math/big"
	"net"
)

func PickIPFromCIDRindex(cidr string, index uint64) (string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR %q: %w", cidr, err)
	}

	maskOnes, bits := ipNet.Mask.Size()
	totalIPs := uint64(1) << (bits - maskOnes)
	if index >= totalIPs {
		return "", fmt.Errorf("index %d out of range for subnet %s", index, ipNet.String())
	}

	baseIP := ipNet.IP.To16()
	if baseIP == nil {
		return "", fmt.Errorf("invalid IP in CIDR %s", cidr)
	}

	ipInt := big.NewInt(0).SetBytes(baseIP)
	ipInt.Add(ipInt, big.NewInt(0).SetUint64(index))

	ipBytes := ipInt.Bytes()
	if len(ipBytes) < net.IPv6len {
		pad := make([]byte, net.IPv6len-len(ipBytes))
		ipBytes = append(pad, ipBytes...)
	}

	var ip net.IP
	if ipNet.IP.To4() != nil {
		ip = net.IP(ipBytes[len(ipBytes)-net.IPv4len:])
	} else {
		ip = net.IP(ipBytes)
	}

	return ip.String(), nil
}

// countIPs returns the total number of IPs in the CIDR range.
// Works for both IPv4 and IPv6.
func CountIPs(ipnet *net.IPNet) uint64 {
	ones, bits := ipnet.Mask.Size()

	// IPv4 can always fit into uint64 directly
	if bits == 32 {
		size := uint64(1) << uint64(bits-ones)
		return size
	}

	// IPv6 may overflow 64 bits → use big.Int
	size := new(big.Int).Lsh(big.NewInt(1), uint(bits-ones))

	// clamp to uint64 maximum if needed
	if size.BitLen() > 64 {
		// you may want to return math.MaxUint64
		// or handle this via conditions
		return ^uint64(0)
	}

	return size.Uint64()
}
