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
	// Create a temporary directory for local file tests
	tempDir, err := os.MkdirTemp("", "fetcher_test_local")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	// Initialize the fetcher with default client
	f := NewFetcher(nil)

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
						t.Errorf("Expected '%s', got '%s'", string(tt.expected), data)
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
						t.Errorf("Expected '%s', got '%s'", string(tt.expected), data)
					}
				}
			}
		})
	}
}

// TestFetcher_URLs tests the Fetch method for URL scenarios.
func TestFetcher_URLs(t *testing.T) {
	// Define table of URL test cases
	urlTests := []struct {
		name          string
		setup         func(t *testing.T) string // Returns the URL to fetch
		expectError   bool
		expected      []byte
		errorContains string
		customClient  func(t *testing.T) *http.Client
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
			customClient: func(_ *testing.T) *http.Client {
				// Create a client with a short timeout
				client := &http.Client{
					Timeout: 1 * time.Nanosecond, // Force timeout
				}
				return client
			},
		},
	}

	for _, tt := range urlTests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			url := tt.setup(t)

			// Instantiate the fetcher with a custom client if provided
			var fetcherInstance Fetcher
			if tt.customClient != nil {
				client := tt.customClient(t)
				fetcherInstance = NewFetcher(client)
			} else {
				fetcherInstance = NewFetcher(nil)
			}

			data, err := fetcherInstance.Fetch("", url)

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
					t.Errorf("Expected '%s', got '%s'", string(tt.expected), data)
				}
			}
		})
	}
}

// TestFetcher_AuthHeader tests that the Authorization header is correctly set when GITHUB_TOKEN is present.
func TestFetcher_AuthHeader(t *testing.T) {
	expectedContent := "Authorized Content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token, exists := os.LookupEnv("GITHUB_TOKEN")
		if exists {
			expectedAuth := fmt.Sprintf("Bearer %s", token)
			if authHeader != expectedAuth {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		} else if authHeader != "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedContent))
	}))
	defer server.Close()

	// Set the GITHUB_TOKEN environment variable
	os.Setenv("GITHUB_TOKEN", "testtoken")
	defer os.Unsetenv("GITHUB_TOKEN")

	// Initialize the fetcher with default client
	f := NewFetcher(nil)

	data, err := f.Fetch("", server.URL)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !bytes.Equal(data, []byte(expectedContent)) {
		t.Errorf("Expected '%s', got '%s'", expectedContent, data)
	}
}

// TestFetcher_NoAuthHeader ensures that no Authorization header is set when GITHUB_TOKEN is absent.
func TestFetcher_NoAuthHeader(t *testing.T) {
	expectedContent := "No Auth Content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedContent))
	}))
	defer server.Close()

	// Ensure GITHUB_TOKEN is not set
	os.Unsetenv("GITHUB_TOKEN")

	// Initialize the fetcher with default client
	f := NewFetcher(nil)

	data, err := f.Fetch("", server.URL)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !bytes.Equal(data, []byte(expectedContent)) {
		t.Errorf("Expected '%s', got '%s'", expectedContent, data)
	}
}
