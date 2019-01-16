package valuesloader

import (
	"fmt"

	"github.com/txgruppi/run/cache"
)

// New returns a new ValuesLoader instance or an error if the provided loader
// function is nil.
func New(loaders ...ValueLoaderFunc) (*ValuesLoader, error) {
	if loaders == nil || len(loaders) == 0 {
		return nil, fmt.Errorf("loaders are required")
	}
	for index, loader := range loaders {
		if loader == nil {
			return nil, fmt.Errorf("nil loader at %d", index)
		}
	}
	return &ValuesLoader{
		loaders: loaders,
		cache:   cache.New(),
	}, nil
}

// ValueLoaderFunc is called whenever the value for a given key is not present
// in the cache.
type ValueLoaderFunc func(key string) (string, bool)

// ValuesLoader loads and caches values based on keys. It gets the values from
// the ValueLoaderFunc provided.
type ValuesLoader struct {
	loaders []ValueLoaderFunc
	cache   cache.Cache
}

// Loopup gets a value for a given key, it will first try to get the value from
// the cache, if it is not present in the cache it will call each
// ValueLoaderFunc in order. It returns true if the key was found, otherwise
// it returns false.
func (v *ValuesLoader) Lookup(key string) (string, bool) {
	if v.cache.Has(key) {
		return v.cache.Get(key), true
	}

	for _, loader := range v.loaders {
		value, ok := loader(key)
		if ok {
			v.cache.Set(key, value)
			return value, true
		}
	}

	return "", false
}

// Get works just like Lookup but without returning the boolean flag.
func (v *ValuesLoader) Get(key string) string {
	value, _ := v.Lookup(key)
	return value
}
