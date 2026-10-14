package utils

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"syscall"
	"time"
)

var (
	// Carrier-grade NAT 100.64.0.0/10
	_, cgnatNet, _ = net.ParseCIDR("100.64.0.0/10")
)

// IsBlockedIP checks if an IP belongs to private, loopback, link-local, carrier-grade NAT, or multicast ranges
func IsBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	// If IPv4-mapped IPv6, convert to IPv4
	if ipv4 := ip.To4(); ipv4 != nil {
		ip = ipv4
	}

	if ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() ||
		ip.IsMulticast() {
		return true
	}

	// Check carrier-grade NAT (100.64.0.0/10)
	if cgnatNet != nil && cgnatNet.Contains(ip) {
		return true
	}

	// Explicit check for Azure/AWS link-local metadata (169.254.169.254)
	if ip.String() == "169.254.169.254" {
		return true
	}

	return false
}

// ValidateURL performs syntactical and host safety validation for input URLs
func ValidateURL(rawURL string, allowPrivate bool) (*url.URL, error) {
	if len(rawURL) == 0 {
		return nil, errors.New("url is required")
	}
	if len(rawURL) > 2048 {
		return nil, errors.New("url exceeds maximum length of 2048 characters")
	}

	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("scheme must be http or https")
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return nil, errors.New("url host is required")
	}

	if !allowPrivate {
		lowerHost := strings.ToLower(hostname)
		if lowerHost == "localhost" || strings.HasSuffix(lowerHost, ".localhost") || strings.HasSuffix(lowerHost, ".local") {
			return nil, errors.New("localhost and local addresses are blocked")
		}

		if ip := net.ParseIP(hostname); ip != nil {
			if IsBlockedIP(ip) {
				return nil, errors.New("private and internal IP addresses are blocked")
			}
		}
	}

	return parsed, nil
}

// NewSafeDialer creates a net.Dialer whose Control function guards against SSRF at connection time
func NewSafeDialer(allowPrivate bool, timeout time.Duration) *net.Dialer {
	return &net.Dialer{
		Timeout: timeout,
		Control: func(network, address string, c syscall.RawConn) error {
			if allowPrivate {
				return nil
			}

			host, _, err := net.SplitHostPort(address)
			if err != nil {
				host = address
			}

			ip := net.ParseIP(host)
			if ip != nil && IsBlockedIP(ip) {
				return fmt.Errorf("connection to private/internal IP (%s) is blocked by SSRF guard", host)
			}

			return nil
		},
	}
}
