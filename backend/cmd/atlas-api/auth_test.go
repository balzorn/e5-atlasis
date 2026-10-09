package main

import (
	"net/http/httptest"
	"testing"
)

func TestPrincipalResolverForConfigDefaultsToDeny(t *testing.T) {
	resolver, err := principalResolverForConfig("", "0.0.0.0:8080")
	if err != nil {
		t.Fatalf("principalResolverForConfig() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/assets/IA00001", nil)
	req.Header.Set("X-Actor-ID", "USR001")
	if _, err := resolver.Resolve(req); err == nil {
		t.Fatal("resolver accepted X-Actor-ID without an explicit development auth mode")
	}
}

func TestDevelopmentHeaderAuthRequiresLoopbackAddress(t *testing.T) {
	for _, addr := range []string{
		":8080",
		"0.0.0.0:8080",
		"192.0.2.10:8080",
		"localhost:8080",
		"not-an-address",
	} {
		t.Run(addr, func(t *testing.T) {
			if _, err := principalResolverForConfig(developmentHeaderAuthMode, addr); err == nil {
				t.Fatalf("principalResolverForConfig(%q) succeeded; want loopback restriction error", addr)
			}
		})
	}

	for _, addr := range []string{"127.0.0.1:8080", "[::1]:8080"} {
		t.Run(addr, func(t *testing.T) {
			if _, err := principalResolverForConfig(developmentHeaderAuthMode, addr); err != nil {
				t.Fatalf("principalResolverForConfig(%q) error = %v", addr, err)
			}
		})
	}
}

func TestPrincipalResolverForConfigRejectsUnknownMode(t *testing.T) {
	if _, err := principalResolverForConfig("trust-any-header", "127.0.0.1:8080"); err == nil {
		t.Fatal("principalResolverForConfig() succeeded for unknown mode")
	}
}
