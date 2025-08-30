package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Load(path string) (*Config, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("accessing path %q: %w", path, err)
	}

	c := DefaultConfig()

	var files []string
	if !info.IsDir() {
		files = []string{path}
	} else {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("reading directory %q: %w", path, err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			ext := filepath.Ext(entry.Name())
			if ext != ".yaml" && ext != ".yml" {
				continue
			}
			files = append(files, filepath.Join(path, entry.Name()))
		}
	}

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("opening file %q: %w", file, err)
		}

		dec := yaml.NewDecoder(f)
		dec.KnownFields(true)
		if err := dec.Decode(c); err != nil {
			f.Close()
			return nil, fmt.Errorf("decoding yaml %q: %w", file, err)
		}
		f.Close()
	}

	return c, nil
}
