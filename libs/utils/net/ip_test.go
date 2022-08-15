package net

import (
	"testing"

	utils "github.com/bananocoin/boompow/libs/utils/testing"
)

func TestIsIPInRange(t *testing.T) {
	ip := "123.45.67.89"

	utils.AssertEqual(t, false, IsIPInHetznerRange(ip))
	utils.AssertEqual(t, true, IsIPInHetznerRange("95.216.77.23"))
}
