package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func Load(dir string) (*Config, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %q: %w", dir, err)
	}

	c := defaultConfig()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("opening file %q: %w", path, err)
		}
		dec := yaml.NewDecoder(f)
		dec.KnownFields(true)
		if err := dec.Decode(c); err != nil {
			f.Close()
			return nil, fmt.Errorf("decoding yaml %q: %w", path, err)
		}
		f.Close()
	}
	return c, nil
}
