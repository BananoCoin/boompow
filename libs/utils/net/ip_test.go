package net

import (
	"testing"

	utils "github.com/bananocoin/boompow/libs/utils/testing"
)

func TestIsIPInRange(t *testing.T) {
	ip := "123.45.67.89"

	utils.AssertEqual(t, false, IsIPInHetznerRange(ip))
	utils.AssertEqual(t, true, IsIPInHetznerRange("95.216.77.23"))
	utils.AssertEqual(t, true, IsIPInHetznerRange("2a01:4f9:c010:780c::1"))
}
