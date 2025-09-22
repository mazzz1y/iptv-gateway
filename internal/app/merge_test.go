package app

import (
	"iptv-gateway/internal/config/proxy"
	"iptv-gateway/internal/config/types"
	"reflect"
	"testing"
)

func TestMergeArrays(t *testing.T) {
	tests := []struct {
		name     string
		arrays   [][]int
		expected []int
	}{
		{
			name:     "empty arrays",
			arrays:   [][]int{},
			expected: []int{},
		},
		{
			name:     "single array",
			arrays:   [][]int{{1, 2, 3}},
			expected: []int{1, 2, 3},
		},
		{
			name:     "multiple arrays",
			arrays:   [][]int{{1, 2}, {3, 4}, {5}},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "empty and non-empty arrays",
			arrays:   [][]int{{1}, {}, {2, 3}},
			expected: []int{1, 2, 3},
		},
		{
			name:     "all empty arrays",
			arrays:   [][]int{{}, {}, {}},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeArrays(tt.arrays...)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("mergeArrays() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMergeArraysStrings(t *testing.T) {
	arrays := [][]string{{"a", "b"}, {"c"}, {"d", "e"}}
	expected := []string{"a", "b", "c", "d", "e"}
	result := mergeArrays(arrays...)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("mergeArrays() = %v, expected %v", result, expected)
	}
}

func TestMergeProxies(t *testing.T) {
	tests := []struct {
		name     string
		proxies  []proxy.Proxy
		expected proxy.Proxy
	}{
		{
			name:     "empty proxies",
			proxies:  []proxy.Proxy{},
			expected: proxy.Proxy{},
		},
		{
			name: "single proxy",
			proxies: []proxy.Proxy{
				{
					Enabled:           boolPtr(true),
					ConcurrentStreams: 5,
				},
			},
			expected: proxy.Proxy{
				Enabled:           boolPtr(true),
				ConcurrentStreams: 5,
			},
		},
		{
			name: "multiple proxies - later overrides",
			proxies: []proxy.Proxy{
				{
					Enabled:           boolPtr(false),
					ConcurrentStreams: 3,
				},
				{
					Enabled:           boolPtr(true),
					ConcurrentStreams: 5,
				},
			},
			expected: proxy.Proxy{
				Enabled:           boolPtr(true),
				ConcurrentStreams: 5,
			},
		},
		{
			name: "zero concurrent streams ignored",
			proxies: []proxy.Proxy{
				{
					ConcurrentStreams: 5,
				},
				{
					ConcurrentStreams: 0,
				},
			},
			expected: proxy.Proxy{
				ConcurrentStreams: 5,
			},
		},
		{
			name: "nil enabled ignored",
			proxies: []proxy.Proxy{
				{
					Enabled: boolPtr(true),
				},
				{
					Enabled: nil,
				},
			},
			expected: proxy.Proxy{
				Enabled: boolPtr(true),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeProxies(tt.proxies...)
			if !reflect.DeepEqual(result.Enabled, tt.expected.Enabled) {
				t.Errorf("mergeProxies().Enabled = %v, expected %v", result.Enabled, tt.expected.Enabled)
			}
			if result.ConcurrentStreams != tt.expected.ConcurrentStreams {
				t.Errorf("mergeProxies().ConcurrentStreams = %v, expected %v", result.ConcurrentStreams, tt.expected.ConcurrentStreams)
			}
		})
	}
}

func TestMergeHandlers(t *testing.T) {
	tests := []struct {
		name     string
		handlers []proxy.Handler
		expected proxy.Handler
	}{
		{
			name:     "empty handlers",
			handlers: []proxy.Handler{},
			expected: proxy.Handler{},
		},
		{
			name: "single handler",
			handlers: []proxy.Handler{
				{
					Command: []string{"cmd1", "arg1"},
					TemplateVars: []types.NameValue{
						{Name: "var1", Value: "value1"},
					},
				},
			},
			expected: proxy.Handler{
				Command: []string{"cmd1", "arg1"},
				TemplateVars: []types.NameValue{
					{Name: "var1", Value: "value1"},
				},
			},
		},
		{
			name: "multiple handlers - command override",
			handlers: []proxy.Handler{
				{
					Command: []string{"cmd1"},
				},
				{
					Command: []string{"cmd2", "arg2"},
				},
			},
			expected: proxy.Handler{
				Command: []string{"cmd2", "arg2"},
			},
		},
		{
			name: "empty command ignored",
			handlers: []proxy.Handler{
				{
					Command: []string{"cmd1"},
				},
				{
					Command: []string{},
				},
			},
			expected: proxy.Handler{
				Command: []string{"cmd1"},
			},
		},
		{
			name: "template vars merged",
			handlers: []proxy.Handler{
				{
					TemplateVars: []types.NameValue{
						{Name: "var1", Value: "value1"},
					},
				},
				{
					TemplateVars: []types.NameValue{
						{Name: "var2", Value: "value2"},
					},
				},
			},
			expected: proxy.Handler{
				TemplateVars: []types.NameValue{
					{Name: "var1", Value: "value1"},
					{Name: "var2", Value: "value2"},
				},
			},
		},
		{
			name: "env vars merged",
			handlers: []proxy.Handler{
				{
					EnvVars: []types.NameValue{
						{Name: "ENV1", Value: "val1"},
					},
				},
				{
					EnvVars: []types.NameValue{
						{Name: "ENV2", Value: "val2"},
					},
				},
			},
			expected: proxy.Handler{
				EnvVars: []types.NameValue{
					{Name: "ENV1", Value: "val1"},
					{Name: "ENV2", Value: "val2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeHandlers(tt.handlers...)
			if !reflect.DeepEqual(result.Command, tt.expected.Command) {
				t.Errorf("mergeHandlers().Command = %v, expected %v", result.Command, tt.expected.Command)
			}
			if !nameValueSlicesEqual(result.TemplateVars, tt.expected.TemplateVars) {
				t.Errorf("mergeHandlers().TemplateVars = %v, expected %v", result.TemplateVars, tt.expected.TemplateVars)
			}
			if !nameValueSlicesEqual(result.EnvVars, tt.expected.EnvVars) {
				t.Errorf("mergeHandlers().EnvVars = %v, expected %v", result.EnvVars, tt.expected.EnvVars)
			}
		})
	}
}

func TestMergePairs(t *testing.T) {
	tests := []struct {
		name     string
		result   []types.NameValue
		handler  []types.NameValue
		expected []types.NameValue
	}{
		{
			name:     "empty handler",
			result:   []types.NameValue{{Name: "var1", Value: "value1"}},
			handler:  []types.NameValue{},
			expected: []types.NameValue{{Name: "var1", Value: "value1"}},
		},
		{
			name:   "empty result",
			result: []types.NameValue{},
			handler: []types.NameValue{
				{Name: "var1", Value: "value1"},
			},
			expected: []types.NameValue{
				{Name: "var1", Value: "value1"},
			},
		},
		{
			name: "merge different variables",
			result: []types.NameValue{
				{Name: "var1", Value: "value1"},
			},
			handler: []types.NameValue{
				{Name: "var2", Value: "value2"},
			},
			expected: []types.NameValue{
				{Name: "var1", Value: "value1"},
				{Name: "var2", Value: "value2"},
			},
		},
		{
			name: "override same variable",
			result: []types.NameValue{
				{Name: "var1", Value: "old_value"},
			},
			handler: []types.NameValue{
				{Name: "var1", Value: "new_value"},
			},
			expected: []types.NameValue{
				{Name: "var1", Value: "new_value"},
			},
		},
		{
			name: "complex merge with override",
			result: []types.NameValue{
				{Name: "var1", Value: "value1"},
				{Name: "var2", Value: "old_value2"},
			},
			handler: []types.NameValue{
				{Name: "var2", Value: "new_value2"},
				{Name: "var3", Value: "value3"},
			},
			expected: []types.NameValue{
				{Name: "var1", Value: "value1"},
				{Name: "var2", Value: "new_value2"},
				{Name: "var3", Value: "value3"},
			},
		},
		{
			name:     "both empty",
			result:   []types.NameValue{},
			handler:  []types.NameValue{},
			expected: []types.NameValue{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := make([]types.NameValue, len(tt.result))
			copy(result, tt.result)

			mergePairs(&result, tt.handler)

			if !nameValueSlicesEqual(result, tt.expected) {
				t.Errorf("mergePairs() result = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func TestUniqueNames(t *testing.T) {
	tests := []struct {
		name     string
		names    []string
		expected []string
	}{
		{
			name:     "empty",
			names:    []string{},
			expected: nil,
		},
		{
			name:     "single names",
			names:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "no duplicates",
			names:    []string{"a", "b", "c", "d", "e"},
			expected: []string{"a", "b", "c", "d", "e"},
		},
		{
			name:     "with duplicates - first occurrence wins",
			names:    []string{"a", "b", "b", "c", "c", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "all duplicates",
			names:    []string{"x", "y", "y", "x", "x"},
			expected: []string{"x", "y"},
		},
		{
			name:     "preset then client pattern",
			names:    []string{"preset1", "preset2", "client1", "preset1"},
			expected: []string{"preset1", "preset2", "client1"},
		},
		{
			name:     "complex scenario",
			names:    []string{"base", "sports", "premium", "base", "movies", "sports", "news"},
			expected: []string{"base", "sports", "premium", "movies", "news"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uniqueNames(tt.names)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("uniqueNames() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func nameValueSlicesEqual(a, b []types.NameValue) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]string, len(a))
	for _, nv := range a {
		aMap[nv.Name] = nv.Value
	}

	bMap := make(map[string]string, len(b))
	for _, nv := range b {
		bMap[nv.Name] = nv.Value
	}

	return reflect.DeepEqual(aMap, bMap)
}
