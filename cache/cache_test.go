package cache_test

import (
	"testing"

	"github.com/nproc/run/cache"
	"github.com/stretchr/testify/assert"
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
