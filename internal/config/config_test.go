package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
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
				if cfg.LogLevel != "info" {
					t.Errorf("expected LogLevel to be 'info', got '%s'", cfg.LogLevel)
				}
				if cfg.Cache.Path != "cache" {
					t.Errorf("expected cache.path to be './cache', got '%s'", cfg.Cache.Path)
				}
			},
		},
		{
			name: "custom values",
			configContent: `
listen_addr: ":9090"
public_url: "https://iptv.example.com"
secret: "custom-secret"
log_level: "debug"
cache:
  path: "/tmp/cache"
  ttl: 48h
`,
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.ListenAddr != ":9090" {
					t.Errorf("expected ListenAddr to be ':9090', got '%s'", cfg.ListenAddr)
				}
				if cfg.PublicURL.String() != "https://iptv.example.com" {
					t.Errorf("expected PublicURL to be 'https://iptv.example.com', got '%s'", cfg.PublicURL.String())
				}
				if cfg.Secret != "custom-secret" {
					t.Errorf("expected Secret to be 'custom-secret', got '%s'", cfg.Secret)
				}
				if cfg.LogLevel != "debug" {
					t.Errorf("expected LogLevel to be 'debug', got '%s'", cfg.LogLevel)
				}
				if cfg.Cache.Path != "/tmp/cache" {
					t.Errorf("expected cache.path to be '/tmp/cache', got '%s'", cfg.Cache.Path)
				}
				if cfg.Cache.TTL != TTL(48*time.Hour) {
					t.Errorf("expected cache.ttl to be 48h, got '%s'", time.Duration(cfg.Cache.TTL))
				}
			},
		},
		{
			name: "with clients and subscriptions",
			configContent: `
listen_addr: ":8080"
public_url: "http://example.com"
secret: "test-secret"
clients:
  - name: "client1"
    secret: "manager-secret"
    subscriptions:
      - "sub1"
      - "sub2"
subscriptions:
  - name: "sub1"
    playlist: "http://example.com/playlist.m3u"
    epg: "http://example.com/epg.xml"
  - name: "sub2"
    playlist: 
      - "http://example.com/playlist1.m3u"
      - "http://example.com/playlist2.m3u"
    epg:
      - "http://example.com/epg1.xml"
      - "http://example.com/epg2.xml"
`,
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.Clients) != 1 {
					t.Fatalf("expected 1 manager, got %d", len(cfg.Clients))
				}
				if cfg.Clients[0].Name != "client1" {
					t.Errorf("expected manager name to be 'client1', got '%s'", cfg.Clients[0].Name)
				}
				if cfg.Clients[0].Secret != "manager-secret" {
					t.Errorf("expected manager secret to be 'manager-secret', got '%s'", cfg.Clients[0].Secret)
				}
				if len(cfg.Clients[0].Subscriptions) != 2 {
					t.Fatalf("expected 2 manager subscriptions, got %d", len(cfg.Clients[0].Subscriptions))
				}
				if cfg.Clients[0].Subscriptions[0] != "sub1" || cfg.Clients[0].Subscriptions[1] != "sub2" {
					t.Errorf("incorrect manager subscriptions: %v", cfg.Clients[0].Subscriptions)
				}

				if len(cfg.Subscriptions) != 2 {
					t.Fatalf("expected 2 subscriptions, got %d", len(cfg.Subscriptions))
				}
				if cfg.Subscriptions[0].Name != "sub1" {
					t.Errorf("expected subscription name to be 'sub1', got '%s'", cfg.Subscriptions[0].Name)
				}
				if len(cfg.Subscriptions[0].Playlist) != 1 || cfg.Subscriptions[0].Playlist[0] != "http://example.com/playlist.m3u" {
					t.Errorf("incorrect playlist URL: %v", cfg.Subscriptions[0].Playlist)
				}
				if len(cfg.Subscriptions[0].EPG) != 1 || cfg.Subscriptions[0].EPG[0] != "http://example.com/epg.xml" {
					t.Errorf("incorrect EPG URL: %v", cfg.Subscriptions[0].EPG)
				}

				if cfg.Subscriptions[1].Name != "sub2" {
					t.Errorf("expected subscription name to be 'sub2', got '%s'", cfg.Subscriptions[1].Name)
				}
				if len(cfg.Subscriptions[1].Playlist) != 2 {
					t.Fatalf("expected 2 playlist URLs, got %d", len(cfg.Subscriptions[1].Playlist))
				}
				if cfg.Subscriptions[1].Playlist[0] != "http://example.com/playlist1.m3u" ||
					cfg.Subscriptions[1].Playlist[1] != "http://example.com/playlist2.m3u" {
					t.Errorf("incorrect playlist URLs: %v", cfg.Subscriptions[1].Playlist)
				}
				if len(cfg.Subscriptions[1].EPG) != 2 {
					t.Fatalf("expected 2 EPG URLs, got %d", len(cfg.Subscriptions[1].EPG))
				}
				if cfg.Subscriptions[1].EPG[0] != "http://example.com/epg1.xml" ||
					cfg.Subscriptions[1].EPG[1] != "http://example.com/epg2.xml" {
					t.Errorf("incorrect EPG URLs: %v", cfg.Subscriptions[1].EPG)
				}
			},
		},
		{
			name: "with proxy configuration",
			configContent: `
listen_addr: ":8080"
public_url: "http://example.com"
secret: "test-secret"
proxy:
  enabled: true
  concurrency: 10
  stream:
    command: ["custom-command", "-i", "{{.url}}", "pipe:1"]
  error:
    command: ["error-command", "-i", "test", "pipe:1"]
    rate_limit_exceeded:
      template_vars:
        message: "Custom rate limit message"
    link_expired:
      template_vars:
        message: "Custom link expired message"
`,
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Proxy.Enabled == nil || *cfg.Proxy.Enabled != true {
					t.Errorf("expected Proxy.Enabled to be true")
				}
				if cfg.Proxy.ConcurrentStreams != 10 {
					t.Errorf("expected Proxy.ConcurrentStreams to be 10, got %d", cfg.Proxy.ConcurrentStreams)
				}
				if len(cfg.Proxy.Stream.Command) != 4 {
					t.Fatalf("expected 4 stream command elements, got %d", len(cfg.Proxy.Stream.Command))
				}
				if cfg.Proxy.Stream.Command[0] != "custom-command" {
					t.Errorf("incorrect stream command: %v", cfg.Proxy.Stream.Command)
				}
				if len(cfg.Proxy.Error.Command) != 4 {
					t.Fatalf("expected 4 error command elements, got %d", len(cfg.Proxy.Error.Command))
				}
				if cfg.Proxy.Error.Command[0] != "error-command" {
					t.Errorf("incorrect error command: %v", cfg.Proxy.Error.Command)
				}
				if cfg.Proxy.Error.RateLimitExceeded.TemplateVars["message"] != "Custom rate limit message" {
					t.Errorf("incorrect rate limit exceeded message: %s",
						cfg.Proxy.Error.RateLimitExceeded.TemplateVars["message"])
				}
				if cfg.Proxy.Error.LinkExpired.TemplateVars["message"] != "Custom link expired message" {
					t.Errorf("incorrect link expired message: %s",
						cfg.Proxy.Error.LinkExpired.TemplateVars["message"])
				}
			},
		},
		{
			name: "with filter configuration",
			configContent: `
listen_addr: ":8080"
public_url: "http://example.com"
secret: "test-secret"
excludes:
  tags:
    group-title: ["Sports", "News"]
  attrs:
    tvg-id: ["sport.*"]
  channel_name: ["BBC.*", "CNN"]
`,
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.Excludes.Tags["group-title"]) != 2 {
					t.Fatalf("expected 2 group-title filters, got %d", len(cfg.Excludes.Tags["group-title"]))
				}
				if cfg.Excludes.Tags["group-title"][0].String() != "Sports" ||
					cfg.Excludes.Tags["group-title"][1].String() != "News" {
					t.Errorf("incorrect group-title filters: %v", cfg.Excludes.Tags["group-title"])
				}

				if len(cfg.Excludes.Attrs["tvg-id"]) != 1 {
					t.Fatalf("expected 1 tvg-id filter, got %d", len(cfg.Excludes.Attrs["tvg-id"]))
				}
				if cfg.Excludes.Attrs["tvg-id"][0].String() != "sport.*" {
					t.Errorf("incorrect tvg-id filter: %v", cfg.Excludes.Attrs["tvg-id"])
				}

				if len(cfg.Excludes.ChannelName) != 2 {
					t.Fatalf("expected 2 channel_name filters, got %d", len(cfg.Excludes.ChannelName))
				}
				if cfg.Excludes.ChannelName[0].String() != "BBC.*" ||
					cfg.Excludes.ChannelName[1].String() != "CNN" {
					t.Errorf("incorrect channel_name filters: %v", cfg.Excludes.ChannelName)
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
			configContent: `listen_addr: {invalid-yaml
`,
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
			name:          "invalid regex in filter",
			configContent: "listen_addr: \":8080\"\npublic_url: \"http://example.com\"\nsecret: \"test-secret\"\nexcludes:\n  channel_name: [\"[*\", \"CNN\"]",
			expectError:   true,
			validate:      nil,
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
