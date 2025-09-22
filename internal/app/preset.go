package app

import (
	"fmt"
	"iptv-gateway/internal/config"
)

type presetResolver struct {
	presetMap       map[string]config.Preset
	clientName      string
	resolvedPresets map[string]bool
}

func newPresetResolver(presets []config.Preset, client config.Client) *presetResolver {
	presetMap := make(map[string]config.Preset, len(presets))
	for _, preset := range presets {
		presetMap[preset.Name] = preset
	}

	return &presetResolver{
		presetMap:       presetMap,
		clientName:      client.Name,
		resolvedPresets: make(map[string]bool),
	}
}

func (r *presetResolver) resolve(presetNames []string) ([]config.Preset, error) {
	var result []config.Preset
	for _, name := range presetNames {
		resolved, err := r.resolveRecursive(name)
		if err != nil {
			return nil, err
		}
		result = append(result, resolved...)
	}
	return result, nil
}

func (r *presetResolver) resolveRecursive(name string) ([]config.Preset, error) {
	if r.resolvedPresets[name] {
		return []config.Preset{}, nil
	}

	preset, exists := r.presetMap[name]
	if !exists {
		return nil, fmt.Errorf("preset '%s' for client '%s' is not defined in config", name, r.clientName)
	}

	r.resolvedPresets[name] = true

	result := []config.Preset{preset}
	for _, childName := range preset.Presets {
		children, err := r.resolveRecursive(childName)
		if err != nil {
			return nil, err
		}
		result = append(result, children...)
	}

	return result, nil
}
