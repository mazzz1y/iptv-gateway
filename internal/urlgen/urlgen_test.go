package urlgen

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name      string
		publicURL string
		secret    string
		wantErr   bool
	}{
		{
			name:      "valid generator",
			publicURL: "https://example.com",
			secret:    "test-secret-key",
			wantErr:   false,
		},
		{
			name:      "empty secret",
			publicURL: "https://example.com",
			secret:    "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := NewGenerator(tt.publicURL, tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("newGenerator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && g == nil {
				t.Error("newGenerator() returned nil generator")
			}
			if !tt.wantErr && g.PublicURL != tt.publicURL {
				t.Errorf("newGenerator() publicURL = %v, want %v", g.PublicURL, tt.publicURL)
			}
		})
	}
}

func TestGenerator_CreateURL(t *testing.T) {
	g, err := NewGenerator("https://example.com", "test-secret")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	tests := []struct {
		name    string
		data    Data
		ttl     time.Duration
		wantExt string
		wantErr bool
	}{
		{
			name: "stream request",
			data: Data{
				RequestType: Stream,
				URL:         "https://stream.example.com/video",
				Playlist:    "playlist1",
				ChannelID:   "channel1",
			},
			ttl:     time.Hour,
			wantExt: ".ts",
			wantErr: false,
		},
		{
			name: "file request with extension",
			data: Data{
				RequestType: File,
				URL:         "https://file.example.com/video.mp4",
				Playlist:    "playlist2",
				ChannelID:   "channel2",
			},
			ttl:     time.Hour,
			wantExt: ".mp4",
			wantErr: false,
		},
		{
			name: "file request without extension",
			data: Data{
				RequestType: File,
				URL:         "https://file.example.com/video",
				Playlist:    "playlist3",
				ChannelID:   "channel3",
			},
			ttl:     time.Hour,
			wantExt: ".file",
			wantErr: false,
		},
		{
			name: "file request with query params",
			data: Data{
				RequestType: File,
				URL:         "https://file.example.com/video.mp4?token=abc123",
				Playlist:    "playlist4",
				ChannelID:   "channel4",
			},
			ttl:     time.Hour,
			wantExt: ".mp4",
			wantErr: false,
		},
		{
			name: "stream request with additional fields",
			data: Data{
				RequestType: Stream,
				URL:         "https://stream.example.com/video",
				Playlist:    "playlist5",
				ChannelID:   "channel5",
			},
			ttl:     time.Hour,
			wantExt: ".ts",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := g.CreateURL(tt.data, tt.ttl)
			if (err != nil) != tt.wantErr {
				t.Errorf("createURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if u == nil {
				t.Error("createURL() returned nil url")
				return
			}

			if !strings.HasPrefix(u.String(), g.PublicURL) {
				t.Errorf("createURL() url doesn't start with public url: %v", u.String())
			}

			if !strings.HasSuffix(u.Path, "/f"+tt.wantExt) {
				t.Errorf("createURL() url doesn't end with expected extension: got %v, want %v", u.Path, "/f"+tt.wantExt)
			}

			parts := strings.Split(u.Path, "/")
			if len(parts) < 2 {
				t.Error("createURL() invalid url format")
				return
			}
			token := parts[len(parts)-2]

			decrypted, err := g.Decrypt(token)
			if err != nil {
				t.Errorf("failed to decrypt token: %v", err)
				return
			}

			if decrypted.RequestType != tt.data.RequestType {
				t.Errorf("decrypted requestType = %v, want %v", decrypted.RequestType, tt.data.RequestType)
			}
			if decrypted.URL != tt.data.URL {
				t.Errorf("decrypted url = %v, want %v", decrypted.URL, tt.data.URL)
			}
			if decrypted.Playlist != tt.data.Playlist {
				t.Errorf("decrypted playlist = %v, want %v", decrypted.Playlist, tt.data.Playlist)
			}
			if decrypted.ChannelID != tt.data.ChannelID {
				t.Errorf("decrypted channelID = %v, want %v", decrypted.ChannelID, tt.data.ChannelID)
			}
		})
	}
}

func TestGenerator_Decrypt(t *testing.T) {
	g, err := NewGenerator("https://example.com", "test-secret")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	validData := Data{
		RequestType: Stream,
		URL:         "https://example.com/video",
		Playlist:    "playlist1",
		ChannelID:   "channel1",
	}

	u, err := g.CreateURL(validData, time.Hour)
	if err != nil {
		t.Fatalf("Failed to create URL: %v", err)
	}

	parts := strings.Split(u.Path, "/")
	validToken := parts[len(parts)-2]

	tests := []struct {
		name    string
		token   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "invalid base64",
			token:   "invalid!@#$%",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "short token",
			token:   "dG9v",
			wantErr: true,
			errMsg:  "invalid token",
		},
		{
			name:    "tampered token",
			token:   validToken[:len(validToken)-4] + "XXXX",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := g.Decrypt(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("decrypt() error = %v, want error containing %v", err, tt.errMsg)
			}
			if !tt.wantErr && data == nil {
				t.Error("decrypt() returned nil data")
			}
		})
	}
}

func TestGenerator_CrossSecretDecryption(t *testing.T) {
	g1, err := NewGenerator("https://example.com", "secret1")
	if err != nil {
		t.Fatalf("failed to create generator1: %v", err)
	}

	g2, err := NewGenerator("https://example.com", "secret2")
	if err != nil {
		t.Fatalf("failed to create generator2: %v", err)
	}

	data := Data{
		RequestType: Stream,
		URL:         "https://example.com/video",
		Playlist:    "playlist1",
		ChannelID:   "channel1",
	}

	u, err := g1.CreateURL(data, time.Hour)
	if err != nil {
		t.Fatalf("Failed to create URL: %v", err)
	}

	parts := strings.Split(u.Path, "/")
	token := parts[len(parts)-2]

	_, err = g2.Decrypt(token)
	if err == nil {
		t.Error("expected error when decrypting with different secret")
	}
}

func TestGenerator_LinkExpiration(t *testing.T) {
	g, err := NewGenerator("https://example.com", "test-secret")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	data := Data{
		RequestType: Stream,
		URL:         "https://example.com/video",
		Playlist:    "playlist1",
		ChannelID:   "channel1",
	}

	tests := []struct {
		name    string
		ttl     time.Duration
		wantErr bool
	}{
		{
			name:    "valid TTL - 1 hour",
			ttl:     time.Hour,
			wantErr: false,
		},
		{
			name:    "valid TTL - 5 minutes",
			ttl:     5 * time.Minute,
			wantErr: false,
		},
		{
			name:    "valid TTL - 24 hours",
			ttl:     24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "zero TTL (no expiration)",
			ttl:     0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := g.CreateURL(data, tt.ttl)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && u == nil {
				t.Error("CreateURL() returned nil url")
			}
		})
	}
}

func TestGenerator_ExpiredLinkDecryption(t *testing.T) {
	g, err := NewGenerator("https://example.com", "test-secret")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	data := Data{
		RequestType: Stream,
		URL:         "https://example.com/video",
		Playlist:    "playlist1",
		ChannelID:   "channel1",
	}

	u, err := g.CreateURL(data, time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create URL: %v", err)
	}

	parts := strings.Split(u.Path, "/")
	token := parts[len(parts)-2]

	time.Sleep(10 * time.Millisecond)

	_, err = g.Decrypt(token)
	if err == nil {
		t.Error("expected error when decrypting expired token")
	}

	if !errors.Is(err, ErrExpiredURL) {
		t.Errorf("expected ErrExpiredURL, got %v", err)
	}
}

func TestGenerator_ExpiredLinkEdgeCases(t *testing.T) {
	g, err := NewGenerator("https://example.com", "test-secret")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	data := Data{
		RequestType: Stream,
		URL:         "https://example.com/video",
		Playlist:    "playlist1",
		ChannelID:   "channel1",
	}

	tests := []struct {
		name        string
		ttl         time.Duration
		sleepTime   time.Duration
		expectError bool
		errorType   error
	}{
		{
			name:        "valid token - no sleep",
			ttl:         time.Second,
			sleepTime:   0,
			expectError: false,
			errorType:   nil,
		},
		{
			name:        "expired token",
			ttl:         10 * time.Millisecond,
			sleepTime:   50 * time.Millisecond,
			expectError: true,
			errorType:   ErrExpiredURL,
		},
		{
			name:        "zero TTL - no expiration",
			ttl:         0,
			sleepTime:   100 * time.Millisecond,
			expectError: false,
			errorType:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := g.CreateURL(data, tt.ttl)
			if err != nil {
				t.Fatalf("Failed to create URL: %v", err)
			}

			parts := strings.Split(u.Path, "/")
			token := parts[len(parts)-2]

			if tt.sleepTime > 0 {
				time.Sleep(tt.sleepTime)
			}

			decryptedData, err := g.Decrypt(token)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if tt.errorType != nil && !errors.Is(err, tt.errorType) {
					t.Errorf("expected error %v, got %v", tt.errorType, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if decryptedData == nil {
					t.Error("expected decrypted data but got nil")
				}
			}
		})
	}
}

func TestDetermineExtension(t *testing.T) {
	tests := []struct {
		name string
		data Data
		want string
	}{
		{
			name: "stream type",
			data: Data{RequestType: Stream},
			want: ".ts",
		},
		{
			name: "file with .mp4",
			data: Data{RequestType: File, URL: "https://example.com/video.mp4"},
			want: ".mp4",
		},
		{
			name: "file with .m3u8",
			data: Data{RequestType: File, URL: "https://example.com/playlist.m3u8"},
			want: ".m3u8",
		},
		{
			name: "file with query params",
			data: Data{RequestType: File, URL: "https://example.com/video.mp4?token=123"},
			want: ".mp4",
		},
		{
			name: "file without extension",
			data: Data{RequestType: File, URL: "https://example.com/video"},
			want: ".file",
		},
		{
			name: "file with multiple dots",
			data: Data{RequestType: File, URL: "https://example.com/my.video.file.mp4"},
			want: ".mp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := determineExtension(tt.data); got != tt.want {
				t.Errorf("determineExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}
