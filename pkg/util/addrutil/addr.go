package addrutil

import "strings"

// ParseAddr is used to parse an address string
func ParseAddr(addr string) string {
	// ip、dns
	if strings.HasPrefix(addr, "ip://") || strings.HasPrefix(addr, "dns://") {
		return strings.Split(addr, "//")[1]
	}
	return addr
}
