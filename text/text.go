package text

import (
	"bytes"
)

// Tokens returns the tokens found in data.
func Tokens(data []byte) []*Token {
	if data == nil {
		return nil
	}
	return parse(data)
}

// Replace replaces a token by a value in data.
func Replace(data []byte, token, value string) []byte {
	if data == nil {
		return nil
	}

	return bytes.Replace(
		data,
		[]byte(token),
		[]byte(value),
		-1,
	)
}

type Token struct {
	Raw  string
	Keys []string
}
