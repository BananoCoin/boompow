package net

import (
	"errors"
	"net"
)

// Some data center ranges we want to block
var hetznerRanges = []string{
	"116.202.0.0/16",
	"116.203.0.0/16",
	"128.140.0.0/17",
	"135.181.0.0/16",
	"136.243.0.0/16",
	"138.201.0.0/16",
	"142.132.128.0/17",
	"144.76.0.0/16",
	"148.251.0.0/16",
	"157.90.0.0/16",
	"159.69.0.0/16",
	"162.55.0.0/16",
	"167.233.0.0/16",
	"167.235.0.0/16",
	"168.119.0.0/16",
	"171.25.225.0/24",
	"176.9.0.0/16",
	"178.212.75.0/24",
	"178.63.0.0/16",
	"185.107.52.0/22",
	"185.110.95.0/24",
	"185.112.180.0/24",
	"185.126.28.0/22",
	"185.12.65.0/24",
	"185.136.140.0/23",
	"185.157.176.0/23",
	"185.157.178.0/23",
	"185.157.83.0/24",
	"185.171.224.0/22",
	"185.189.228.0/24",
	"185.189.229.0/24",
	"185.189.230.0/24",
	"185.189.231.0/24",
	"185.209.124.0/22",
	"185.213.45.0/24",
	"185.216.237.0/24",
	"185.226.99.0/24",
	"185.228.8.0/23",
	"185.242.76.0/24",
	"185.36.144.0/22",
	"185.50.120.0/23",
	"188.34.128.0/17",
	"188.40.0.0/16",
	"193.110.6.0/23",
	"193.163.198.0/24",
	"193.25.170.0/23",
	"194.35.12.0/23",
	"194.42.180.0/22",
	"194.42.184.0/22",
	"194.62.106.0/24",
	"195.201.0.0/16",
	"195.248.224.0/24",
	"195.60.226.0/24",
	"195.96.156.0/24",
	"197.242.84.0/22",
	"201.131.3.0/24",
	"213.133.96.0/19",
	"213.232.193.0/24",
	"213.239.192.0/18",
	"23.88.0.0/17",
	"45.148.28.0/22",
	"45.15.120.0/22",
	"46.4.0.0/16",
	"49.12.0.0/16",
	"49.13.0.0/16",
	"5.75.128.0/17",
	"5.9.0.0/16",
	"78.46.0.0/15",
	"83.219.100.0/22",
	"83.243.120.0/22",
	"85.10.192.0/18",
	"88.198.0.0/16",
	"88.99.0.0/16",
	"91.107.128.0/17",
	"91.190.240.0/21",
	"91.233.8.0/22",
	"94.130.0.0/16",
	"94.154.121.0/24",
	"95.217.0.0/16",
	"95.216.0.0/16",
	"65.21.0.0/16",
	"65.109.0.0/16",
	"65.108.0.0/16",
	"45.136.70.0/23",
	"135.181.0.0/16",
	"2a01:4f8::/32",
	"2a01:4f9::/32",
	"2a01:4ff:ff01::/48",
	"2a01:b140::/29",
	"2a06:1301:4050::/48",
	"2a06:be80::/29",
	"2a0e:2c80::/29",
	"2a11:48c0::/29",
	"2a11:e980::/29",
	"2a12:e00::/29",
}

type IPMatcher struct {
	IP     net.IP
	SubNet *net.IPNet
}
type IPMatchers []*IPMatcher

func NewIPMatcher(ipStr string) (*IPMatcher, error) {
	ip, subNet, err := net.ParseCIDR(ipStr)
	if err != nil {
		ip = net.ParseIP(ipStr)
		if ip == nil {
			return nil, errors.New("invalid IP: " + ipStr)
		}
	}
	return &IPMatcher{ip, subNet}, nil
}

func (m IPMatcher) Match(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	return m.IP.Equal(ip) || m.SubNet != nil && m.SubNet.Contains(ip)
}

func NewIPMatchers(ips []string) (list IPMatchers, err error) {
	for _, ipStr := range ips {
		var m *IPMatcher
		m, err = NewIPMatcher(ipStr)
		if err != nil {
			return
		}
		list = append(list, m)
	}
	return
}

func IPContains(ipMatchers []*IPMatcher, ip string) bool {
	for _, m := range ipMatchers {
		if m.Match(ip) {
			return true
		}
	}
	return false
}

func IsIPInHetznerRange(ip string) bool {
	for _, rangeStr := range hetznerRanges {
		matcher, err := NewIPMatcher(rangeStr)
		if err != nil {
			return true
		}
		if matcher.Match(ip) {
			return true
		}
	}
	return false
}
