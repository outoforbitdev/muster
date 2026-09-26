package cmd

import "testing"

func TestWrapText(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		width int
		want  []string
	}{
		{"empty string", "", 10, []string{""}},
		{"fits on one line", "short text", 20, []string{"short text"}},
		{"wraps on word boundaries", "the quick brown fox jumps", 10, []string{"the quick", "brown fox", "jumps"}},
		{"zero width returns unwrapped", "some text", 0, []string{"some text"}},
		{"single long word exceeding width is not split", "supercalifragilisticexpialidocious", 10, []string{"supercalifragilisticexpialidocious"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.s, tt.width)
			if len(got) != len(tt.want) {
				t.Fatalf("wrapText(%q, %d) = %v, want %v", tt.s, tt.width, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("wrapText(%q, %d) = %v, want %v", tt.s, tt.width, got, tt.want)
				}
			}
		})
	}
}
