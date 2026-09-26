package cmd

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// descriptionWrapWidth is the max line length for wrapped description text
// in table output.
const descriptionWrapWidth = 50

// renderTable writes headers and rows as an aligned, tab-separated table.
// Column widths are derived from content, so rows must already contain
// their final display text (e.g. pre-wrapped description lines).
func renderTable(w io.Writer, headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, strings.Join(headers, "\t"))
	for _, row := range rows {
		_, _ = fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	_ = tw.Flush()
}

// wrapText splits s into lines of at most width runes, breaking on word
// boundaries. An empty string returns a single empty line, so callers can
// always render at least one row per entry.
func wrapText(s string, width int) []string {
	if s == "" {
		return []string{""}
	}
	if width <= 0 {
		return []string{s}
	}

	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	line := words[0]
	for _, word := range words[1:] {
		if len(line)+1+len(word) > width {
			lines = append(lines, line)
			line = word
			continue
		}
		line += " " + word
	}
	lines = append(lines, line)

	return lines
}
