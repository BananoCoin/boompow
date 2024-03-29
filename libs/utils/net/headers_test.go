package net

import (
	"bytes"
	"net/http"
	"testing"

	utils "github.com/bananocoin/boompow/libs/utils/testing"
)

func TestGetIPAddressFromHeader(t *testing.T) {
	ip := "123.45.67.89"

	// 4 methods of getting IP Address, CF-Connecting-IP preferred, X-Real-Ip, then X-Forwarded-For, then RemoteAddr

	request, _ := http.NewRequest(http.MethodPost, "appditto.com", bytes.NewReader([]byte("")))
	request.Header.Set("CF-Connecting-IP", ip)
	request.Header.Set("X-Real-Ip", "not-the-ip")
	request.Header.Set("X-Forwarded-For", "not-the-ip")
	utils.AssertEqual(t, ip, GetIPAddress(request))

	request, _ = http.NewRequest(http.MethodPost, "appditto.com", bytes.NewReader([]byte("")))
	request.Header.Set("X-Real-Ip", ip)
	request.Header.Set("X-Forwarded-For", "not-the-ip")

	utils.AssertEqual(t, ip, GetIPAddress(request))

	request, _ = http.NewRequest(http.MethodPost, "appditto.com", bytes.NewReader([]byte("")))
	request.Header.Set("X-Forwarded-For", ip)
	utils.AssertEqual(t, ip, GetIPAddress(request))

	request, _ = http.NewRequest(http.MethodPost, "appditto.com", bytes.NewReader([]byte("")))
	request.RemoteAddr = ip
	utils.AssertEqual(t, ip, GetIPAddress(request))
}
