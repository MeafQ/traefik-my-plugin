package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/traefik/traefik/v3/pkg/server/middleware"
)

type Config struct {
	DisallowedIPs    []string `json:"disallowedIPs,omitempty"`
	AllowedSubnet    string   `json:"allowedSubnet,omitempty"`
	RequiredHeader   string   `json:"requiredHeader,omitempty"`
	RequiredValue    string   `json:"requiredValue,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		DisallowedIPs:    []string{},
		AllowedSubnet:    "10.0.0.0/8",
		RequiredHeader:   "X-Custom-Header",
		RequiredValue:    "ExpectedValue",
	}
}

type AccessControl struct {
	next            http.Handler
	name            string
	disallowedIPs   map[string]struct{}
	allowedSubnet   string
	requiredHeader  string
	requiredValue   string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	disallowed := make(map[string]struct{})
	for _, ip := range config.DisallowedIPs {
		disallowed[ip] = struct{}{}
	}
	return &AccessControl{
		next:            next,
		name:            name,
		disallowedIPs:   disallowed,
		allowedSubnet:   config.AllowedSubnet,
		requiredHeader:  config.RequiredHeader,
		requiredValue:   config.RequiredValue,
	}, nil
}

func (a *AccessControl) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	clientIP := req.RemoteAddr
	if idx := strings.LastIndex(clientIP, ":"); idx != -1 {
		clientIP = clientIP[:idx] // Remove port if present
	}

	// Check disallowed IPs
	if _, exists := a.disallowedIPs[clientIP]; exists {
		http.Error(rw, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if allowed by subnet
	// TODO: Implement subnet check

	// Check required header
	if value := req.Header.Get(a.requiredHeader); value == a.requiredValue {
		a.next.ServeHTTP(rw, req)
		return
	}

	http.Error(rw, "Forbidden", http.StatusForbidden)
}
