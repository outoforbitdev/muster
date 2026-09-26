package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/outoforbitdev/muster/internal/workspace"
)

const (
	ansiReset = "\x1b[0m"
	ansiGreen = "\x1b[32m"
	ansiRed   = "\x1b[31m"
)

// colorEnabled reports whether ANSI colors should be emitted, honoring the
// NO_COLOR convention (https://no-color.org/): any non-empty value disables
// color.
func colorEnabled() bool {
	return os.Getenv("NO_COLOR") == ""
}

// colorize wraps text in the given ANSI code, resetting afterward. Empty
// text is returned as-is, and coloring is skipped entirely when disabled.
func colorize(text, code string) string {
	if text == "" || !colorEnabled() {
		return text
	}
	return code + text + ansiReset
}

// green colorizes text for display as safe to clean.
func green(text string) string {
	return colorize(text, ansiGreen)
}

// red colorizes text for display as not safe to clean.
func red(text string) string {
	return colorize(text, ansiRed)
}

// padColored right-pads text with spaces to width (based on text's visible
// length, ignoring the ANSI codes added afterward), then colorizes it. The
// padding is applied to the plain text first so tabwriter-style column
// alignment survives colorizing, since ANSI escape codes are invisible but
// would otherwise be counted by naive width calculations.
func padColored(text string, width int, code string) string {
	padding := width - len(text)
	if padding < 0 {
		padding = 0
	}
	return colorize(text, code) + strings.Repeat(" ", padding)
}

// cleanRow is one row of the "clean" command's WORKSPACE/STACK/REPO/STATE
// table. workspaceColor colors the workspace and stack cells (based on
// whether the whole workspace is safe to clean); repoColor colors the repo
// and state cells (based on that repo's own status).
type cleanRow struct {
	workspace, stack, repo, state string
	workspaceColor, repoColor     string
}

// renderCleanTable writes headers and colored, column-aligned rows. Column
// widths are fixed (not measured from content, unlike renderTable) because
// callers must size them from the same plain-text lengths used to decide
// how STATE text wraps, before any ANSI codes are added; a naive
// tabwriter-style measurement would otherwise count the invisible escape
// bytes as visible width and misalign columns.
func renderCleanTable(w io.Writer, workspaceWidth, stackWidth, repoWidth int, rows []cleanRow) {
	_, _ = fmt.Fprintf(w, "%s  %s  %s  %s\n",
		padPlain("WORKSPACE", workspaceWidth),
		padPlain("STACK", stackWidth),
		padPlain("REPO", repoWidth),
		"STATE",
	)

	for _, row := range rows {
		_, _ = fmt.Fprintf(w, "%s  %s  %s  %s\n",
			padColored(row.workspace, workspaceWidth, row.workspaceColor),
			padColored(row.stack, stackWidth, row.workspaceColor),
			padColored(row.repo, repoWidth, row.repoColor),
			colorize(row.state, row.repoColor),
		)
	}
}

// repoColorFor maps a repo's status to the color its REPO/STATE cells
// should render in: green for verified clean, red for anything that blocks
// cleaning, and no color for a skipped (non-git) directory, since it's
// neither confirmed safe nor a risk to ignore.
func repoColorFor(status string) string {
	switch status {
	case workspace.StatusClean:
		return ansiGreen
	case workspace.StatusSkipped:
		return ""
	default:
		return ansiRed
	}
}

// padPlain right-pads text with spaces to width, with no coloring.
func padPlain(text string, width int) string {
	padding := width - len(text)
	if padding < 0 {
		padding = 0
	}
	return text + strings.Repeat(" ", padding)
}
