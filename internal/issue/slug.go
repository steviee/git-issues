package issue

import (
	"fmt"
	"regexp"
	"strings"
)

var umlautMap = map[rune]string{
	'ä': "ae", 'ö': "oe", 'ü': "ue", 'ß': "ss",
	'Ä': "ae", 'Ö': "oe", 'Ü': "ue",
}

var allowedChars = regexp.MustCompile(`[^a-z0-9 -]`)
var multiDash = regexp.MustCompile(`-{2,}`)

func GenerateSlug(title string) string {
	s := strings.ToLower(title)

	// Replace umlauts
	var buf strings.Builder
	for _, r := range s {
		if repl, ok := umlautMap[r]; ok {
			buf.WriteString(repl)
		} else {
			buf.WriteRune(r)
		}
	}
	s = buf.String()

	// Remove everything except a-z, 0-9, space, dash
	s = allowedChars.ReplaceAllString(s, "")

	// Replace spaces with dashes
	s = strings.ReplaceAll(s, " ", "-")

	// Collapse multiple dashes
	s = multiDash.ReplaceAllString(s, "-")

	// Truncate to 40 characters
	if len(s) > 40 {
		s = s[:40]
	}

	// Strip trailing dash
	s = strings.TrimRight(s, "-")

	return s
}

func Filename(id int, slug string) string {
	return fmt.Sprintf("%04d-%s.md", id, slug)
}
