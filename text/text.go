package text

import (
	"bytes"
	"regexp"
)

var (
	tokenRegexp *regexp.Regexp
)

func init() {
	tokenRegexp = regexp.MustCompile(`(?i)\{\{([A-Z0-9_-]+)\}\}`)
}

// Tokens returns the tokens found in data.
// A token is defined as `(?i)\{\{([A-Z0-9_-]+)\}\}`.
func Tokens(data []byte) []string {
	if data == nil {
		return nil
	}

	tokens := []string{}
	known := map[string]struct{}{}

	matches := tokenRegexp.FindAllSubmatch(data, -1)

	for _, match := range matches {
		token := string(match[1])
		if _, ok := known[token]; ok {
			continue
		}
		known[token] = struct{}{}
		tokens = append(tokens, token)
	}

	return tokens
}

// Replace replaces a token by a value in data.
func Replace(data []byte, token, value string) []byte {
	if data == nil {
		return nil
	}

	return bytes.Replace(
		data,
		[]byte("{{"+token+"}}"),
		[]byte(value),
		-1,
	)
}
