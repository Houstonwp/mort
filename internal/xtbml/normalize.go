package xtbml

import (
	"strings"
	"unicode"
)

// NormalizeIdentifier converts human-readable table names into a predictable slug.
func NormalizeIdentifier(input string) string {
	var b strings.Builder
	lastUnderscore := false

	for _, r := range strings.ToLower(input) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastUnderscore = false
		case r == '_' || unicode.IsSpace(r) || r == '-' || r == '/':
			if !lastUnderscore && b.Len() > 0 {
				b.WriteByte('_')
				lastUnderscore = true
			}
		default:
			// Skip other punctuation.
		}
	}

	out := b.String()
	out = strings.Trim(out, "_")
	out = strings.ReplaceAll(out, "__", "_")
	return out
}
