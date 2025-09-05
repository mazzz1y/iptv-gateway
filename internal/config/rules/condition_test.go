package rules

import (
	"iptv-gateway/internal/config/types"
	"regexp"
	"testing"
)

func TestConditionWithNamedReference(t *testing.T) {
	testCond := Condition{
		Ref: "adult",
	}

	if testCond.IsEmpty() {
		t.Error("Condition with Ref should not be empty")
	}

	namedCondition := NamedCondition{
		Name: "adult",
		When: []Condition{
			{
				Tag: &TagCondition{
					Name:  "EXTGRP",
					Value: types.RegexpArr{regexp.MustCompile("(?i)adult")},
				},
			},
		},
	}

	if namedCondition.Name != "adult" {
		t.Errorf("Expected named condition name to be 'adult', got %s", namedCondition.Name)
	}

	if len(namedCondition.When) != 1 {
		t.Errorf("Expected 1 condition in When, got %d", len(namedCondition.When))
	}

	if namedCondition.When[0].Tag.Name != "EXTGRP" {
		t.Errorf("Expected tag name to be 'EXTGRP', got %s", namedCondition.When[0].Tag.Name)
	}
}

func TestConditionIsEmpty(t *testing.T) {
	tests := []struct {
		name      string
		condition Condition
		expected  bool
	}{
		{
			name:      "Empty condition",
			condition: Condition{},
			expected:  true,
		},
		{
			name: "Condition with Ref",
			condition: Condition{
				Ref: "test",
			},
			expected: false,
		},
		{
			name: "Condition with Name",
			condition: Condition{
				Name: types.RegexpArr{regexp.MustCompile("test")},
			},
			expected: false,
		},
		{
			name: "Condition with Tag",
			condition: Condition{
				Tag: &TagCondition{
					Name:  "test",
					Value: types.RegexpArr{regexp.MustCompile("value")},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.condition.IsEmpty(); got != tt.expected {
				t.Errorf("Condition.IsEmpty() = %v, want %v", got, tt.expected)
			}
		})
	}
}
