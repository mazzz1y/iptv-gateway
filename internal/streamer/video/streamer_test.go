package video

import (
	"strings"
	"testing"
)

func TestStreamer_PrepareCommand_WithTemplateVars(t *testing.T) {
	command := []string{"cmd", "{{.Param1}}", "--option={{.Param2}}"}
	envVars := map[string]string{}
	templateVars := map[string]any{
		"Param1": "value1",
		"Param2": "value2",
	}

	streamer, err := NewStreamer(command, envVars, templateVars)
	if err != nil {
		t.Fatalf("Unexpected error creating streamer: %v", err)
	}

	commandParts, err := streamer.renderCommand(templateVars)

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
	command := []string{"cmd", "{{.Param1"}
	envVars := map[string]string{}
	templateVars := map[string]any{}

	_, err := NewStreamer(command, envVars, templateVars)
	if err == nil {
		t.Fatal("Expected error for invalid template, got nil")
	}
}

func TestStreamer_PrepareCommand_WithNestedVars(t *testing.T) {
	command := []string{"ffmpeg", "-i", "{{.Source}}"}
	envVars := map[string]string{}
	templateVars := map[string]any{
		"Source":   "{{.Protocol}}://{{.Server}}/{{.Path}}",
		"Protocol": "rtmp",
		"Server":   "example.com",
		"Path":     "live/stream",
	}

	streamer, err := NewStreamer(command, envVars, templateVars)
	if err != nil {
		t.Fatalf("Unexpected error creating streamer: %v", err)
	}

	commandParts, err := streamer.renderCommand(templateVars)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedParts := []string{"ffmpeg", "-i", "{{.Protocol}}://{{.Server}}/{{.Path}}"}
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
	command := []string{
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
	}
	envVars := map[string]string{}
	templateVars := map[string]any{
		"Width":      1280,
		"Height":     720,
		"Resolution": 720,
		"Preset":     "veryfast",
	}

	streamer, err := NewStreamer(command, envVars, templateVars)
	if err != nil {
		t.Fatalf("Unexpected error creating streamer: %v", err)
	}

	commandParts, err := streamer.renderCommand(templateVars)

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
