package utils

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestGenerateShortID(t *testing.T) {
	const attempts = 1000
	seen := make(map[string]bool)

	for i := 0; i < attempts; i++ {
		id, err := GenerateShortID()
		if err != nil {
			t.Fatalf("unexpected error on attempt %d: %v", i, err)
		}

		if len(id) != ShortIDLength {
			t.Errorf("expected ID length %d, got %d", ShortIDLength, len(id))
		}

		if seen[id] {
			t.Errorf("duplicate ID found on attempt %d: %s", i, id)
		}

		seen[id] = true
	}
}

func TestGetBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		req      *http.Request
		expected string
	}{
		{
			name: "plain HTTP",
			req: &http.Request{
				Host:   "example.com",
				Header: http.Header{},
			},
			expected: "http://example.com/",
		},
		{
			name: "HTTPS via TLS",
			req: &http.Request{
				Host: "secure.com",
				TLS:  &tls.ConnectionState{},
			},
			expected: "https://secure.com/",
		},
		{
			name: "HTTPS via X-Forwarded-Proto",
			req: &http.Request{
				Host: "proxy.com",
				Header: http.Header{
					"X-Forwarded-Proto": []string{"https"},
				},
			},
			expected: "https://proxy.com/",
		},
		{
			name: "X-Forwarded-Proto is http",
			req: &http.Request{
				Host: "proxied-http.com",
				Header: http.Header{
					"X-Forwarded-Proto": []string{"http"},
				},
			},
			expected: "http://proxied-http.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBaseURL(tt.req)
			if got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}
