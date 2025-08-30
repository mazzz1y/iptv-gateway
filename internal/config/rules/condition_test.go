package rules

import (
	"testing"

	"iptv-gateway/internal/config/types"
)

func TestCondition_IsEmpty(t *testing.T) {
	tests := []struct {
		name      string
		condition Condition
		expected  bool
	}{
		{
			name:      "empty condition",
			condition: Condition{},
			expected:  true,
		},
		{
			name: "condition with name",
			condition: Condition{
				Name: types.RegexpArr{},
			},
			expected: true,
		},
		{
			name: "condition with attribute",
			condition: Condition{
				Attr: &AttributeCondition{
					Name:  "test",
					Value: types.RegexpArr{},
				},
			},
			expected: false,
		},
		{
			name: "condition with tag",
			condition: Condition{
				Tag: &TagCondition{
					Name:  "test",
					Value: types.RegexpArr{},
				},
			},
			expected: false,
		},
		{
			name: "condition with And",
			condition: Condition{
				And: []Condition{{}},
			},
			expected: false,
		},
		{
			name: "condition with Or",
			condition: Condition{
				Or: []Condition{{}},
			},
			expected: false,
		},
		{
			name: "condition with Not",
			condition: Condition{
				Not: []Condition{{}},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.condition.IsEmpty()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
