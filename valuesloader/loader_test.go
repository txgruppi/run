package valuesloader_test

import (
	"testing"

	"github.com/nproc/run/valuesloader"
	"github.com/stretchr/testify/assert"
)

func TestValuesLoader(t *testing.T) {
	env := map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
		"d": "4",
	}

	loaderFunc := func(key string) string {
		return env[key]
	}

	t.Run("nil loader", func(t *testing.T) {
		assert := assert.New(t)

		loader, err := valuesloader.New(nil)
		assert.Nil(loader)
		assert.EqualError(err, "loader is required")
	})

	t.Run("valid execution", func(t *testing.T) {
		assert := assert.New(t)

		loader, err := valuesloader.New(loaderFunc)
		assert.Nil(err)
		assert.NotNil(loader)

		for key, value := range env {
			assert.Equal(value, loader.Get(key))
		}
	})
}
