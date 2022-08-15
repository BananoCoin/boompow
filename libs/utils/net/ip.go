package net

import (
	"net"
)

// Some data center ranges we want to block
var hetznerRanges = []string{
	"95.217.0.0/16",
	"95.216.0.0/16",
	"65.21.0.0/16",
	"65.109.0.0/16",
	"65.108.0.0/16",
	"45.136.70.0/23",
	"135.181.0.0/16",
}

func IsIPInHetznerRange(ip string) bool {
	for _, rangeStr := range hetznerRanges {
		_, subnet, _ := net.ParseCIDR(rangeStr)
		ip := net.ParseIP(ip)
		if subnet.Contains(ip) {
			return true
		}

	}
	return false
}
