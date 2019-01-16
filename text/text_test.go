package text_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/txgruppi/run/text"
)

func TestText(t *testing.T) {
	data := []byte(`[database]
	url = "{{MONGO_URL}}"

	[jwt]
	secret = "{{ JWT_SECRET   }}"

	[server]
	bind = "{{ server.bind || SERVER_BIND}}"
	port = "{{server.port||SERVER_PORT}}"

	[other_server]
	bind = "{{   server.bind ||   SERVER_BIND   }}"
	port = "{{server.por||SERVER_PORT}}"
	`)
	expectedTokens := []*text.Token{
		&text.Token{
			Raw:  "{{MONGO_URL}}",
			Keys: []string{"MONGO_URL"},
		},
		&text.Token{
			Raw:  "{{ JWT_SECRET   }}",
			Keys: []string{"JWT_SECRET"},
		},
		&text.Token{
			Raw:  "{{ server.bind || SERVER_BIND}}",
			Keys: []string{"server.bind", "SERVER_BIND"},
		},
		&text.Token{
			Raw:  "{{server.port||SERVER_PORT}}",
			Keys: []string{"server.port", "SERVER_PORT"},
		},
		&text.Token{
			Raw:  "{{   server.bind ||   SERVER_BIND   }}",
			Keys: []string{"server.bind", "SERVER_BIND"},
		},
		&text.Token{
			Raw:  "{{server.por||SERVER_PORT}}",
			Keys: []string{"server.por", "SERVER_PORT"},
		},
	}
	expectedData := []byte(`[database]
	url = "0"

	[jwt]
	secret = "1"

	[server]
	bind = "2"
	port = "3"

	[other_server]
	bind = "4"
	port = "5"
	`)

	t.Run("nil data", func(t *testing.T) {
		assert := assert.New(t)

		assert.Nil(text.Tokens(nil))
		assert.Nil(text.Replace(nil, "k", "v"))
	})

	t.Run("valid execution", func(t *testing.T) {
		assert := assert.New(t)

		tokens := text.Tokens(data)
		assert.NotNil(tokens)

		assert.Equal(expectedTokens, tokens)

		actual := data
		for index, token := range tokens {
			actual = text.Replace(actual, token.Raw, strconv.Itoa(index))
		}
		assert.Equal(expectedData, actual)
	})
}
