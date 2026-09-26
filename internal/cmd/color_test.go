package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func withNoColorEnv(t *testing.T, value string) {
	t.Helper()
	orig, had := os.LookupEnv("NO_COLOR")
	if err := os.Setenv("NO_COLOR", value); err != nil {
		t.Fatalf("failed to set NO_COLOR: %v", err)
	}
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("NO_COLOR", orig)
		} else {
			_ = os.Unsetenv("NO_COLOR")
		}
	})
}

func withoutNoColorEnv(t *testing.T) {
	t.Helper()
	orig, had := os.LookupEnv("NO_COLOR")
	_ = os.Unsetenv("NO_COLOR")
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("NO_COLOR", orig)
		}
	})
}

func TestGreenWrapsTextInAnsiCodes(t *testing.T) {
	withoutNoColorEnv(t)

	got := green("safe")
	want := "\x1b[32msafe\x1b[0m"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestRedWrapsTextInAnsiCodes(t *testing.T) {
	withoutNoColorEnv(t)

	got := red("unsafe")
	want := "\x1b[31munsafe\x1b[0m"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestColorizeEmptyStringStaysEmpty(t *testing.T) {
	withoutNoColorEnv(t)

	if got := green(""); got != "" {
		t.Errorf("expected empty string to stay empty, got %q", got)
	}
}

func TestNoColorEnvDisablesColor(t *testing.T) {
	withNoColorEnv(t, "1")

	if got := green("safe"); got != "safe" {
		t.Errorf("expected NO_COLOR to disable coloring, got %q", got)
	}
}

func TestPadColoredPadsBeforeColorReset(t *testing.T) {
	withoutNoColorEnv(t)

	got := padColored("abc", 6, ansiGreen)
	want := "\x1b[32mabc\x1b[0m   "
	if got != want {
		t.Errorf("expected padding after the reset code, got %q", got)
	}
}

func TestPadColoredEmptyTextStaysPlain(t *testing.T) {
	withoutNoColorEnv(t)

	got := padColored("", 6, ansiGreen)
	want := "      "
	if got != want {
		t.Errorf("expected plain spaces for empty text, got %q", got)
	}
}

func TestPadColoredWithNoColorEnv(t *testing.T) {
	withNoColorEnv(t, "1")

	got := padColored("abc", 6, ansiGreen)
	want := "abc   "
	if got != want {
		t.Errorf("expected NO_COLOR to disable coloring while still padding, got %q", got)
	}
}

func TestRenderCleanTableColorsCellsPerRow(t *testing.T) {
	withoutNoColorEnv(t)

	rows := []cleanRow{
		{workspace: "ws1", stack: "", repo: "repo-a", state: "clean, merged", workspaceColor: ansiGreen, repoColor: ansiGreen},
		{workspace: "", stack: "", repo: "repo-b", state: "dirty", workspaceColor: ansiGreen, repoColor: ansiRed},
	}

	var buf bytes.Buffer
	renderCleanTable(&buf, 9, 5, 6, rows)

	got := buf.String()
	wantHeader := "WORKSPACE  STACK  REPO    STATE\n"
	if !strings.HasPrefix(got, wantHeader) {
		t.Fatalf("expected header %q, got %q", wantHeader, got)
	}

	wantRow1 := "\x1b[32mws1\x1b[0m      " + "  " + "     " + "  " + "\x1b[32mrepo-a\x1b[0m" + "  " + "\x1b[32mclean, merged\x1b[0m\n"
	wantRow2 := "         " + "  " + "     " + "  " + "\x1b[31mrepo-b\x1b[0m" + "  " + "\x1b[31mdirty\x1b[0m\n"

	if !strings.Contains(got, wantRow1) {
		t.Errorf("expected output to contain colored row 1 %q, got %q", wantRow1, got)
	}
	if !strings.Contains(got, wantRow2) {
		t.Errorf("expected output to contain colored row 2 %q, got %q", wantRow2, got)
	}
}
