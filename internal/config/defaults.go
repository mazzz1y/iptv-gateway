package config

import (
	"iptv-gateway/internal/config/proxy"
	"iptv-gateway/internal/config/types"
	"net/url"
	"time"
)

func DefaultConfig() *Config {
	publicUrl, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	return &Config{
		Server: ServerConfig{
			ListenAddr: ":8080",
			PublicURL:  types.URL(*publicUrl),
		},
		Logs: Logs{
			"info",
			"text",
		},
		URLGenerator: URLGeneratorConfig{
			Secret:    "xxx",
			StreamTTL: types.Duration(30 * 24 * time.Hour),
			FileTTL:   types.Duration(0),
		},
		Cache: CacheConfig{
			Path:        "cache",
			TTL:         types.Duration(24 * time.Hour),
			Retention:   types.Duration(24 * time.Hour * 30),
			Compression: false,
		},
		Proxy: proxy.Proxy{
			Stream: proxy.Handler{
				Command: types.StringOrArr{
					"ffmpeg",
					"-v", "{{ default \"fatal\" .ffmpeg_log_level }}",
					"-i", "{{.url}}",
					"-c", "copy",
					"-f", "mpegts",
					"pipe:1",
				},
				TemplateVars: []types.NameValue{
					{Name: "ffmpeg_log_level", Value: "fatal"},
				},
			},
			Error: proxy.Error{
				Handler: proxy.Handler{
					Command: types.StringOrArr{
						"ffmpeg",
						"-v", "{{ default \"fatal\" .ffmpeg_log_level }}",
						"-f", "lavfi",
						"-i", "color=#301934:size=1280x720:rate=1",
						"-vf", "drawtext=text='{{.message}}':fontcolor=white:fontsize=36:x=(w-text_w)/2:y=(h-text_h)/2+(line_h/2):text_align=C+M",
						"-c:v", "libx264",
						"-preset", "ultrafast",
						"-tune", "stillimage",
						"-g", "1",
						"-r", "1",
						"-t", "15",
						"-pix_fmt", "yuv420p",
						"-f", "mpegts",
						"pipe:1",
					},
					TemplateVars: []types.NameValue{
						{Name: "ffmpeg_log_level", Value: "fatal"},
					},
				},
				RateLimitExceeded: proxy.Handler{
					TemplateVars: []types.NameValue{
						{Name: "message", Value: "Rate limit exceeded\n\nPlease try again later"},
					},
				},
				LinkExpired: proxy.Handler{
					TemplateVars: []types.NameValue{
						{Name: "message", Value: "Link has expired\n\nPlease refresh your playlist"},
					},
				},
				UpstreamError: proxy.Handler{
					TemplateVars: []types.NameValue{
						{Name: "message", Value: "Unable to play stream\n\nPlease try again later or contact administrator"},
					},
				},
			},
		},
	}
}
