package config

import (
	"testing"
	"time"

	"iptv-gateway/internal/config/types"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ListenAddr != ":8080" {
		t.Errorf("expected default ListenAddr to be ':8080', got '%s'", cfg.ListenAddr)
	}

	if cfg.Log.Level != "info" {
		t.Errorf("expected default LogLevel to be 'info', got '%s'", cfg.Log.Level)
	}

	if cfg.Cache.Path != "cache" {
		t.Errorf("expected default cache path to be 'cache', got '%s'", cfg.Cache.Path)
	}

	expectedTTL := types.Duration(24 * time.Hour)
	if cfg.Cache.TTL != expectedTTL {
		t.Errorf("expected default cache TTL to be %v, got %v", expectedTTL, cfg.Cache.TTL)
	}

	expectedRetention := types.Duration(24 * time.Hour * 30)
	if cfg.Cache.Retention != expectedRetention {
		t.Errorf("expected default cache retention to be %v, got %v", expectedRetention, cfg.Cache.Retention)
	}

	if len(cfg.Proxy.Stream.Command) == 0 {
		t.Error("expected default proxy stream command to be set")
	}

	if cfg.Proxy.Stream.Command[0] != "ffmpeg" {
		t.Errorf("expected default proxy stream command to start with 'ffmpeg', got '%s'", cfg.Proxy.Stream.Command[0])
	}

	if len(cfg.Proxy.Error.Command) == 0 {
		t.Error("expected default proxy error command to be set")
	}

	if cfg.Proxy.Error.RateLimitExceeded.TemplateVars == nil {
		t.Error("expected default rate limit exceeded template vars to be set")
	}

	if cfg.Proxy.Error.LinkExpired.TemplateVars == nil {
		t.Error("expected default link expired template vars to be set")
	}

	if cfg.Proxy.Error.UpstreamError.TemplateVars == nil {
		t.Error("expected default upstream error template vars to be set")
	}
}
