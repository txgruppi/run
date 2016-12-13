package cache

// New returns a new Cache instance.
func New() Cache {
	return map[string]string{}
}

// Cache is a map of keys and values meant to cache values for
// valuesloader.ValuesLoader.
type Cache map[string]string

// Has returns if a given key exists in the cache.
func (c Cache) Has(key string) bool {
	_, ok := c[key]
	return ok
}

// Set sets the value for a given key.
func (c Cache) Set(key, value string) {
	c[key] = value
}

// Get returns the value for a given key or an empty string is the key is not
// set.
func (c Cache) Get(key string) string {
	return c[key]
}
