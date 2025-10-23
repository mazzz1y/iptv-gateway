package playlist

import (
	"sort"
	"testing"
)

func TestNaturalLess(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{
			name: "numbers - less than",
			a:    "2",
			b:    "10",
			want: true,
		},
		{
			name: "numbers - equal",
			a:    "5",
			b:    "5",
			want: false,
		},
		{
			name: "channel names with numbers",
			a:    "Channel 2",
			b:    "Channel 10",
			want: true,
		},
		{
			name: "alphabetic comparison",
			a:    "abc",
			b:    "def",
			want: true,
		},
		{
			name: "number before text",
			a:    "123",
			b:    "abc",
			want: true,
		},
		{
			name: "empty strings",
			a:    "",
			b:    "",
			want: false,
		},
		{
			name: "first empty",
			a:    "",
			b:    "abc",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := naturalLess(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("naturalLess(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestExtractNumber(t *testing.T) {
	tests := []struct {
		input      string
		wantNum    int64
		wantLength int
	}{
		{"5", 5, 1},
		{"123", 123, 3},
		{"42abc", 42, 2},
		{"abc", 0, 0},
		{"007", 7, 3},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotNum, gotLength := extractNumber(tt.input)
			if gotNum != tt.wantNum || gotLength != tt.wantLength {
				t.Errorf("extractNumber(%q) = (%v, %v), want (%v, %v)",
					tt.input, gotNum, gotLength, tt.wantNum, tt.wantLength)
			}
		})
	}
}

func TestNaturalSort(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "Channels",
			input: []string{"HBO 2", "HBO 10", "CNN", "CNN HD"},
			want:  []string{"CNN", "CNN HD", "HBO 2", "HBO 10"},
		},
		{
			name:  "numbers only",
			input: []string{"100", "20", "3"},
			want:  []string{"3", "20", "100"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make([]string, len(tt.input))
			copy(got, tt.input)

			sort.Slice(got, func(i, j int) bool {
				return naturalLess(got[i], got[j])
			})

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got %v, want %v", got, tt.want)
					break
				}
			}
		})
	}
}
