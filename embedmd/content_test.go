// content_test.go
package embedmd

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestFetcher_LocalFiles tests the Fetch method for local file scenarios.
func TestFetcher_LocalFiles(t *testing.T) {
	// Initialize the fetcher
	var f Fetcher = fetcher{}

	// Create a temporary directory for local file tests
	tempDir, err := os.MkdirTemp("", "fetcher_test_local")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	// Define table of local file test cases
	localTests := []struct {
		name        string
		setup       func() string // Returns the path to fetch
		expectError bool
		expected    []byte
	}{
		{
			name: "fetch_existing_file",
			setup: func() string {
				fileName := "testfile.txt"
				fileContent := "Hello, World!"
				filePath := filepath.Join(tempDir, fileName)
				if err := os.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
					t.Fatalf("Failed to write temp file: %v", err)
				}
				return fileName
			},
			expectError: false,
			expected:    []byte("Hello, World!"),
		},
		{
			name: "fetch_non_existent_file",
			setup: func() string {
				return "nonexistent.txt"
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "fetch_with_absolute_path",
			setup: func() string {
				fileName := "absolute_testfile.txt"
				fileContent := "Absolute Path Content"
				absFilePath := filepath.Join(tempDir, fileName)
				if err := os.WriteFile(absFilePath, []byte(fileContent), 0644); err != nil {
					t.Fatalf("Failed to write absolute path file: %v", err)
				}
				return absFilePath // Absolute path
			},
			expectError: false,
			expected:    []byte("Absolute Path Content"),
		},
	}

	for _, tt := range localTests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			// Determine if the path is absolute
			if !filepath.IsAbs(path) {
				// Relative path; use tempDir as base
				data, err := f.Fetch(tempDir, path)
				if tt.expectError {
					if err == nil {
						t.Errorf("Expected error, got nil")
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
					if !bytes.Equal(data, tt.expected) {
						t.Errorf("Expected '%s', got '%s'", string(tt.expected), string(data))
					}
				}
			} else {
				// Absolute path; base directory should be ignored
				data, err := f.Fetch("/any/base/dir", path)
				if tt.expectError {
					if err == nil {
						t.Errorf("Expected error, got nil")
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
					if !bytes.Equal(data, tt.expected) {
						t.Errorf("Expected '%s', got '%s'", string(tt.expected), string(data))
					}
				}
			}
		})
	}
}

// TestFetcher_URLs tests the Fetch method for URL scenarios.
func TestFetcher_URLs(t *testing.T) {
	// Initialize the fetcher
	var f Fetcher = fetcher{}

	// Define table of URL test cases
	urlTests := []struct {
		name          string
		setup         func(t *testing.T) string // Returns the URL to fetch
		expectError   bool
		expected      []byte
		errorContains string
		customClient  func()
	}{
		{
			name: "fetch_valid_url",
			setup: func(t *testing.T) string {
				expectedContent := "Mock HTTP Content"
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(expectedContent)) //nolint:errcheck
				}))
				t.Cleanup(server.Close)
				return server.URL
			},
			expectError: false,
			expected:    []byte("Mock HTTP Content"),
		},
		{
			name: "fetch_url_non_ok_status",
			setup: func(t *testing.T) string {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.NotFound(w, r)
				}))
				t.Cleanup(server.Close)
				return server.URL
			},
			expectError:   true,
			expected:      nil,
			errorContains: "404",
		},
		{
			name: "fetch_url_error",
			setup: func(t *testing.T) string {
				return "http://invalid.url"
			},
			expectError:   true,
			expected:      nil,
			errorContains: "no such host",
		},
		{
			name: "fetch_url_empty_response_body",
			setup: func(t *testing.T) string {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				t.Cleanup(server.Close)
				return server.URL
			},
			expectError: false,
			expected:    []byte{},
		},
		{
			name: "fetch_url_redirect",
			setup: func(t *testing.T) string {
				redirectTarget := "Final Content"
				server := httptest.NewServer(nil)
				t.Cleanup(server.Close)

				redirectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Redirect(w, r, server.URL+"/final", http.StatusFound)
				})

				finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(redirectTarget)) //nolint:errcheck
				})

				mux := http.NewServeMux()
				mux.HandleFunc("/", redirectHandler)
				mux.HandleFunc("/final", finalHandler)
				server.Config.Handler = mux

				return server.URL
			},
			expectError: false,
			expected:    []byte("Final Content"),
		},
		{
			name: "fetch_url_timeout",
			setup: func(t *testing.T) string {
				// Create a listener that doesn't accept connections to simulate timeout
				ln, err := net.Listen("tcp", "127.0.0.1:0")
				if err != nil {
					t.Fatalf("Failed to create listener: %v", err)
				}
				addr := ln.Addr().String()
				ln.Close() // Close immediately to make the URL invalid
				return fmt.Sprintf("http://%s", addr)
			},
			expectError:   true,
			expected:      nil,
			errorContains: "timeout",
			customClient: func() {
				// Override the HTTP client with a timeout
				originalClient := http.DefaultClient
				http.DefaultClient = &http.Client{
					Timeout: 1 * time.Nanosecond, // Force timeout
				}
				// Restore the original client after the test
				// Note: `t.Cleanup` ensures this runs even if the test fails
				t.Cleanup(func() { http.DefaultClient = originalClient })
			},
		},
	}

	for _, tt := range urlTests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			url := tt.setup(t)

			// Apply any custom client settings
			if tt.customClient != nil {
				tt.customClient()
			}

			data, err := f.Fetch("", url)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if tt.errorContains != "" && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errorContains)) {
					t.Errorf("Expected error to contain '%s', got '%v'", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !bytes.Equal(data, tt.expected) {
					t.Errorf("Expected '%s', got '%s'", string(tt.expected), string(data))
				}
			}
		})
	}
}
