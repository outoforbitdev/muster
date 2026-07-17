package workspace

import "strings"

// SubstituteTemplate replaces {workspace} placeholders in a branch template.
// Example: "feature-{workspace}" with workspace "my-feature" becomes "feature-my-feature".
func SubstituteTemplate(template, workspace string) string {
	return strings.ReplaceAll(template, "{workspace}", workspace)
}
