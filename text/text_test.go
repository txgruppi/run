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
	secret = "{{JWT_SECRET}}"

	[server]
	bind = "{{SERVER_BIND}}"
	port = "{{SERVER_PORT}}"

	[to_test_repeated_tokens]
	bind = "{{SERVER_BIND}}"
	port = "{{SERVER_PORT}}"
	`)
	expectedTokens := []string{"MONGO_URL", "JWT_SECRET", "SERVER_BIND", "SERVER_PORT"}
	expectedData := []byte(`[database]
	url = "0"

	[jwt]
	secret = "1"

	[server]
	bind = "2"
	port = "3"

	[to_test_repeated_tokens]
	bind = "2"
	port = "3"
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
			actual = text.Replace(actual, token, strconv.Itoa(index))
		}
		assert.Equal(expectedData, actual)
	})
}
