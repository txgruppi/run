package valuesloader

import (
	"fmt"

	"github.com/nproc/run/cache"
)

// New returns a new ValuesLoader instance or an error if the provided loader
// function is nil.
func New(loader ValueLoaderFunc) (*ValuesLoader, error) {
	if loader == nil {
		return nil, fmt.Errorf("loader is required")
	}
	return &ValuesLoader{
		loader: loader,
		cache:  cache.New(),
	}, nil
}

// ValueLoaderFunc is called whenever the value for a given key is not present
// in the cache.
type ValueLoaderFunc func(key string) string

// ValuesLoader loads and caches values based on keys. It gets the values from
// the ValueLoaderFunc provided.
type ValuesLoader struct {
	loader ValueLoaderFunc
	cache  cache.Cache
}

// Get gets a value for a given key, it will first try to get the value from the
// cache, if it is not present in the cache it will call the ValueLoaderFunc.
func (v *ValuesLoader) Get(key string) string {
	if !v.cache.Has(key) {
		v.cache.Set(key, v.loader(key))
	}
	return v.cache.Get(key)
}
