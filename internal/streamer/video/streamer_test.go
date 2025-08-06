package video

import (
	"strings"
	"testing"
)

func TestStreamer_PrepareCommand_EmptyCommand(t *testing.T) {
	streamer := NewStreamer(StreamerConfig{Command: []string{}})
	_, err := streamer.prepareCommand()

	if err == nil {
		t.Fatal("Expected error for empty command, got nil")
	}

	if !strings.Contains(err.Error(), "command cannot be empty") {
		t.Fatalf("Expected 'command cannot be empty' error, got: %v", err)
	}
}

func TestStreamer_PrepareCommand_WithTemplateVars(t *testing.T) {
	config := StreamerConfig{
		Command: []string{"cmd", "{{.Param1}}", "--option={{.Param2}}"},
		TemplateVars: map[string]any{
			"Param1": "value1",
			"Param2": "value2",
		},
	}

	streamer := NewStreamer(config)
	commandParts, err := streamer.prepareCommand()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedParts := []string{"cmd", "value1", "--option=value2"}
	if len(commandParts) != len(expectedParts) {
		t.Fatalf("Expected %d command parts, got %d", len(expectedParts), len(commandParts))
	}

	for i, expected := range expectedParts {
		if commandParts[i] != expected {
			t.Fatalf("Expected command part '%s', got '%s'", expected, commandParts[i])
		}
	}
}

func TestStreamer_PrepareCommand_InvalidTemplate(t *testing.T) {
	config := StreamerConfig{
		Command: []string{"cmd", "{{.Param1"},
	}

	streamer := NewStreamer(config)
	_, err := streamer.prepareCommand()

	if err == nil {
		t.Fatal("Expected error for invalid template, got nil")
	}
}

func TestStreamer_PrepareCommand_WithNestedVars(t *testing.T) {
	config := StreamerConfig{
		Command: []string{"ffmpeg", "-i", "{{.Source}}"},
		TemplateVars: map[string]any{
			"Source":   "{{.Protocol}}://{{.Server}}/{{.Path}}",
			"Protocol": "rtmp",
			"Server":   "example.com",
			"Path":     "live/stream",
		},
	}

	streamer := NewStreamer(config)
	commandParts, err := streamer.prepareCommand()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedParts := []string{"ffmpeg", "-i", "rtmp://example.com/live/stream"}
	if len(commandParts) != len(expectedParts) {
		t.Fatalf("Expected %d command parts, got %d", len(expectedParts), len(commandParts))
	}

	for i, expected := range expectedParts {
		if commandParts[i] != expected {
			t.Fatalf("Expected command part '%s', got '%s'", expected, commandParts[i])
		}
	}
}

func TestStreamer_PrepareCommand_WithNumericVars(t *testing.T) {
	config := StreamerConfig{
		Command: []string{
			`ffmpeg -i input.mp4 -vf scale={{.Width}}:{{.Height}}
{{if lt .Resolution 720}}
-global_quality 24
{{else if lt .Resolution 1080}}
-global_quality 23
{{else if lt .Resolution 2160}}
-global_quality 22
{{else}}
-global_quality 20
{{end}}
-preset {{.Preset}}`,
		},
		TemplateVars: map[string]any{
			"Width":      1280,
			"Height":     720,
			"Resolution": 720,
			"Preset":     "veryfast",
		},
	}

	streamer := NewStreamer(config)
	commandParts, err := streamer.prepareCommand()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedParts := []string{
		"sh",
		"-c",
		"ffmpeg -i input.mp4 -vf scale=1280:720 \n\n-global_quality 23\n\n-preset veryfast",
	}

	if len(commandParts) != len(expectedParts) {
		t.Fatalf("Expected %d command parts, got %d", len(expectedParts), len(commandParts))
	}

	for i, expected := range expectedParts {
		normalizedExpected := strings.ReplaceAll(expected, " ", "")
		normalizedActual := strings.ReplaceAll(commandParts[i], " ", "")

		if normalizedActual != normalizedExpected {
			t.Fatalf("Expected command part '%s', got '%s'", expected, commandParts[i])
		}
	}
}
