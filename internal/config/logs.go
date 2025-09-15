package config

import (
	"fmt"
	"strings"
)

type Logs struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func (l *Logs) Validate() error {
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	if l.Level != "" {
		found := false
		for _, level := range validLevels {
			if l.Level == level {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf(
				"invalid log level '%s', must be one of: %s", l.Level, strings.Join(validLevels, ", "))
		}
	}

	validFormats := []string{"text", "json"}
	if l.Format != "" {
		found := false
		for _, format := range validFormats {
			if l.Format == format {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf(
				"invalid log format '%s', must be one of: %s", l.Format, strings.Join(validFormats, ", "))
		}
	}
	return nil
}
