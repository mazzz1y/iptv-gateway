package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		expectError   bool
		validate      func(t *testing.T, cfg *Config)
	}{
		{
			name: "valid minimal config",
			configContent: `listen_addr: ":8080"
public_url: "http://example.com"
secret: "test-secret"`,
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.ListenAddr != ":8080" {
					t.Errorf("expected ListenAddr to be ':8080', got '%s'", cfg.ListenAddr)
				}
				if cfg.PublicURL.String() != "http://example.com" {
					t.Errorf("expected PublicURL to be 'http://example.com', got '%s'", cfg.PublicURL.String())
				}
				if cfg.Secret != "test-secret" {
					t.Errorf("expected Secret to be 'test-secret', got '%s'", cfg.Secret)
				}
			},
		},
		{
			name:          "file not found",
			configContent: "",
			expectError:   true,
			validate:      nil,
		},
		{
			name: "invalid yaml",
			configContent: `listen_addr: {invalid-yaml`,
			expectError: true,
			validate:    nil,
		},
		{
			name:          "invalid public URL",
			configContent: "listen_addr: \":8080\"\npublic_url: \"://invalid\"\nsecret: \"test-secret\"",
			expectError:   true,
			validate:      nil,
		},
		{
			name: "directory with multiple files",
			configContent: `listen_addr: ":8080"
public_url: "http://example.com"
secret: "test-secret"`,
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.ListenAddr != ":8080" {
					t.Errorf("expected ListenAddr to be ':8080', got '%s'", cfg.ListenAddr)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "config-test")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			var configPath string
			if tt.configContent != "" {
				configDir := filepath.Join(tmpDir, "config")
				err = os.MkdirAll(configDir, 0755)
				if err != nil {
					t.Fatalf("failed to create config directory: %v", err)
				}

				configPath = configDir
				configFile := filepath.Join(configDir, "config.yaml")
				err = os.WriteFile(configFile, []byte(tt.configContent), 0644)
				if err != nil {
					t.Fatalf("failed to write config file: %v", err)
				}
			} else {
				configPath = filepath.Join(tmpDir, "nonexistent")
			}

			cfg, err := Load(configPath)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}
