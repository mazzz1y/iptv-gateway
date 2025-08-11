package urlgen

import (
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
			wantExt: ".mp4",
			wantErr: false,
		},
		{
			name: "with custom created time",
			data: Data{
				RequestType: Stream,
				URL:         "https://stream.example.com/video",
				Playlist:    "playlist5",
				ChannelID:   "channel5",
				Created:     time.Now().Add(-time.Hour),
			},
			wantExt: ".ts",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := g.CreateURL(tt.data)
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

			// Extract token from URL
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
		Created:     time.Now(),
	}

	u, err := g.CreateURL(validData)
	if err != nil {
		t.Fatalf("Failed to create URL: %v", err)
	}

	// Extract token from URL
	parts := strings.Split(u.Path, "/")
	validToken := parts[len(parts)-2]

	expiredData := Data{
		RequestType: File,
		URL:         "https://example.com/old",
		Playlist:    "old-playlist",
		ChannelID:   "old-channel",
		Created:     time.Now().Add(-32 * 24 * time.Hour), // 32 days ago
	}

	expiredURL, err := g.CreateURL(expiredData)
	if err != nil {
		t.Fatalf("failed to create expired url: %v", err)
	}
	expiredParts := strings.Split(expiredURL.Path, "/")
	expiredToken := expiredParts[len(expiredParts)-2]

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
			name:    "expired token",
			token:   expiredToken,
			wantErr: true,
			errMsg:  "expired URL",
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

	u, err := g1.CreateURL(data)
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

func BenchmarkGenerator_CreateURL(b *testing.B) {
	g, err := NewGenerator("https://example.com", "test-secret")
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}

	data := Data{
		RequestType: Stream,
		URL:         "https://example.com/video",
		Playlist:    "playlist1",
		ChannelID:   "channel1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := g.CreateURL(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerator_Decrypt(b *testing.B) {
	g, err := NewGenerator("https://example.com", "test-secret")
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}

	data := Data{
		RequestType: Stream,
		URL:         "https://example.com/video",
		Playlist:    "playlist1",
		ChannelID:   "channel1",
	}

	u, err := g.CreateURL(data)
	if err != nil {
		b.Fatalf("Failed to create URL: %v", err)
	}

	parts := strings.Split(u.Path, "/")
	token := parts[len(parts)-2]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := g.Decrypt(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}
