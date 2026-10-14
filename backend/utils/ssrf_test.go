package utils

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIsBlockedIP(t *testing.T) {
	blockedCases := []string{
		"127.0.0.1",
		"127.0.0.2",
		"10.0.0.1",
		"172.16.0.1",
		"172.31.255.255",
		"192.168.1.1",
		"169.254.169.254", // Cloud metadata
		"100.64.0.1",      // CGNAT
		"0.0.0.0",
		"::1",
		"fc00::1",
		"fe80::1",
	}

	for _, ipStr := range blockedCases {
		ip := net.ParseIP(ipStr)
		if !IsBlockedIP(ip) {
			t.Errorf("expected IP %s to be blocked", ipStr)
		}
	}

	allowedCases := []string{
		"8.8.8.8",
		"1.1.1.1",
		"93.184.216.34", // example.com
	}

	for _, ipStr := range allowedCases {
		ip := net.ParseIP(ipStr)
		if IsBlockedIP(ip) {
			t.Errorf("expected IP %s to be allowed", ipStr)
		}
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		url          string
		allowPrivate bool
		wantErr      bool
	}{
		{"https://example.com", false, false},
		{"http://1.1.1.1", false, false},
		{"http://localhost", false, true},
		{"http://localhost:8080", false, true},
		{"http://test.localhost", false, true},
		{"http://127.0.0.1:8080", false, true},
		{"http://10.0.0.1", false, true},
		{"http://169.254.169.254/latest/meta-data/", false, true},
		{"ftp://example.com", false, true},
		{"http://localhost", true, false},
		{"http://127.0.0.1:8080", true, false},
	}

	for _, tt := range tests {
		_, err := ValidateURL(tt.url, tt.allowPrivate)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateURL(%q, allowPrivate=%v) error = %v, wantErr %v", tt.url, tt.allowPrivate, err, tt.wantErr)
		}
	}
}

func TestSafeDialerBlocksPrivateConnection(t *testing.T) {
	// Start a local test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	dialer := NewSafeDialer(false, 2*time.Second)
	transport := &http.Transport{
		DialContext: dialer.DialContext,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   3 * time.Second,
	}

	// ts.URL points to 127.0.0.1, so safe dialer should block it
	_, err := client.Get(ts.URL)
	if err == nil {
		t.Errorf("expected connection to %s to be blocked, but succeeded", ts.URL)
	}

	// Now allow private
	allowedDialer := NewSafeDialer(true, 2*time.Second)
	allowedClient := &http.Client{
		Transport: &http.Transport{DialContext: allowedDialer.DialContext},
		Timeout:   3 * time.Second,
	}
	resp, err := allowedClient.Get(ts.URL)
	if err != nil {
		t.Errorf("expected connection with allowPrivate=true to succeed, got %v", err)
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func TestSafeDialerBlocksRedirectToPrivate(t *testing.T) {
	// A local server that redirects to 127.0.0.1
	internalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer internalServer.Close()

	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, internalServer.URL, http.StatusFound)
	}))
	defer redirectServer.Close()

	dialer := NewSafeDialer(false, 2*time.Second)
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
		Timeout: 3 * time.Second,
	}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodHead, redirectServer.URL, nil)
	_, err := client.Do(req)
	if err == nil {
		t.Errorf("expected redirect to private IP to be blocked, but succeeded")
	}
}
