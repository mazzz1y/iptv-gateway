package config

import "time"

func defaultConfig() *Config {
	return &Config{
		ListenAddr: ":8080",
		LogLevel:   "info",
		Cache: CacheConfig{
			Path: "cache",
			TTL:  TTL(30 * 24 * time.Hour),
		},
		Proxy: Proxy{
			Stream: Handler{
				Command: []string{"ffmpeg", "-i", "{{.url}}", "-c", "copy", "-f", "mpegts", "pipe:1"},
			},
			Error: Error{
				Handler: Handler{
					Command: []string{
						"ffmpeg",
						"-f", "lavfi",
						"-i", "smptebars=size=1280x720:rate=1",
						"-vf", "drawtext=text='{{.message}}':fontcolor=white:fontsize=36:x=(w-text_w)/2:y=(h-text_h)/2:box=1:boxcolor=black@0.5:boxborderw=10",
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
				},
				RateLimitExceeded: Handler{
					TemplateVars: map[string]string{
						"message": "Rate limit exceeded. Please try again later.",
					},
				},
				LinkExpired: Handler{
					TemplateVars: map[string]string{
						"message": "Link has expired. Please refresh your playlist.",
					},
				},
				UpstreamError: Handler{
					TemplateVars: map[string]string{
						"message": "Unable to play stream. Please try again later or contact administrator.",
					},
				},
			},
		},
	}
}
