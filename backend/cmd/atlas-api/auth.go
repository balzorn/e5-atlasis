package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/balzorn/e5-atlasis/backend/internal/httpapi"
)

const developmentHeaderAuthMode = "development-header"

func principalResolverForConfig(mode, addr string) (httpapi.PrincipalResolver, error) {
	switch strings.TrimSpace(mode) {
	case "":
		return httpapi.NewRejectingPrincipalResolver(), nil
	case developmentHeaderAuthMode:
		if !isLoopbackAddress(addr) {
			return nil, fmt.Errorf("%q authentication mode requires HTTP_ADDR to use a loopback IP address", developmentHeaderAuthMode)
		}
		return httpapi.NewDevelopmentHeaderPrincipalResolver(), nil
	default:
		return nil, fmt.Errorf("unsupported ATLASIS_AUTH_MODE %q", mode)
	}
}

func isLoopbackAddress(addr string) bool {
	host, _, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil {
		return false
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
