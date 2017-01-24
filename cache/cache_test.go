package cache_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/txgruppi/run/cache"
)

func TestCache(t *testing.T) {
	assert := assert.New(t)

	pairs := map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
		"d": "4",
		"e": "",
		"":  "6",
	}

	c := cache.New()

	for key, value := range pairs {
		assert.False(c.Has(key))
		c.Set(key, value)
		assert.True(c.Has(key))
		assert.Equal(value, c.Get(key))
	}
}
