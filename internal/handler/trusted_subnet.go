package handler

import (
	"net"
	"net/http"
	"strings"
)

func trustedSubnetAllowed(trustedSubnet string, r *http.Request) bool {
	if strings.TrimSpace(trustedSubnet) == "" {
		return false
	}
	ip := net.ParseIP(strings.TrimSpace(r.Header.Get("X-Real-IP")))
	if ip == nil {
		return false
	}
	_, network, err := net.ParseCIDR(strings.TrimSpace(trustedSubnet))
	if err != nil || !network.Contains(ip) {
		return false
	}
	return true
}
