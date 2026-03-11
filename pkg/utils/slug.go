package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func GenerateSlug(title string) string {

	if title == "" {
		return ""
	}

	// 1. Handle accents (e.g., 'é' becomes 'e')
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	slug, _, _ := transform.String(t, title)

	// 2. To Lowercase
	slug = strings.ToLower(slug)

	// 3. Replace anything that isn't a lowercase letter or number with a hyphen
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")

	// 4. Trim hyphens from the start and end (e.g., "-my-slug-" -> "my-slug")
	slug = strings.Trim(slug, "-")

	return slug
}
