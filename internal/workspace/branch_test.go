package workspace

import (
	"testing"
)

func TestSubstituteTemplate(t *testing.T) {
	tests := []struct {
		template  string
		workspace string
		expected  string
	}{
		{"feature-{workspace}", "my-feature", "feature-my-feature"},
		{"{workspace}", "test", "test"},
		{"main", "test", "main"},
		{"feature-{workspace}-v1", "my-feature", "feature-my-feature-v1"},
		{"", "test", ""},
		{"feature-{workspace}", "", "feature-"},
		{"{workspace}-{workspace}", "test", "test-test"},
	}

	for _, tt := range tests {
		t.Run(tt.template+"/"+tt.workspace, func(t *testing.T) {
			result := SubstituteTemplate(tt.template, tt.workspace)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
