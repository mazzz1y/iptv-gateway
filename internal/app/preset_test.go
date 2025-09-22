package app

import (
	"iptv-gateway/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolvePresets_Empty(t *testing.T) {
	var presets []config.Preset
	client := config.Client{Name: "test-client"}
	resolver := newPresetResolver(presets, client)

	result, err := resolver.resolve([]string{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestResolvePresets_NotFound(t *testing.T) {
	var presets []config.Preset
	client := config.Client{Name: "test-client"}
	resolver := newPresetResolver(presets, client)

	_, err := resolver.resolve([]string{"missing"})
	assert.Error(t, err)
	assert.Equal(t, "preset 'missing' for client 'test-client' is not defined in config", err.Error())
}

func TestResolvePresets_SimplePreset(t *testing.T) {
	presets := []config.Preset{{Name: "basic"}}
	client := config.Client{Name: "test-client"}
	resolver := newPresetResolver(presets, client)

	result, err := resolver.resolve([]string{"basic"})
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "basic", result[0].Name)
}

func TestResolvePresets_NestedPresets(t *testing.T) {
	presets := []config.Preset{
		{Name: "parent", Presets: []string{"child"}},
		{Name: "child"},
	}
	client := config.Client{Name: "test-client"}
	resolver := newPresetResolver(presets, client)

	result, err := resolver.resolve([]string{"parent"})
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "parent", result[0].Name)
	assert.Equal(t, "child", result[1].Name)
}

func TestResolvePresets_CircularDependency(t *testing.T) {
	presets := []config.Preset{
		{Name: "preset1", Presets: []string{"preset2"}},
		{Name: "preset2", Presets: []string{"preset1"}},
	}
	client := config.Client{Name: "test-client"}
	resolver := newPresetResolver(presets, client)

	result, err := resolver.resolve([]string{"preset1"})
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestResolvePresets_DiamondDependency(t *testing.T) {
	presets := []config.Preset{
		{Name: "root", Presets: []string{"branchA", "branchB"}},
		{Name: "branchA", Presets: []string{"common"}},
		{Name: "branchB", Presets: []string{"common"}},
		{Name: "common"},
	}
	client := config.Client{Name: "test-client"}
	resolver := newPresetResolver(presets, client)

	result, err := resolver.resolve([]string{"root"})
	assert.NoError(t, err)
	assert.Len(t, result, 4)
	expected := []string{"root", "branchA", "common", "branchB"}
	for i, preset := range result {
		assert.Equal(t, expected[i], preset.Name)
	}
}

func TestResolvePresets_MultipleRoots(t *testing.T) {
	presets := []config.Preset{
		{Name: "preset1", Presets: []string{"child"}},
		{Name: "preset2"},
		{Name: "child"},
	}
	client := config.Client{Name: "test-client"}
	resolver := newPresetResolver(presets, client)

	result, err := resolver.resolve([]string{"preset1", "preset2"})
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	expected := []string{"preset1", "child", "preset2"}
	for i, preset := range result {
		assert.Equal(t, expected[i], preset.Name)
	}
}
