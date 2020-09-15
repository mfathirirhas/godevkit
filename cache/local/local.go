package local

import (
	"time"

	_localCache "github.com/patrickmn/go-cache"
)

const (
	defaultCleanupInterval = 1 // every 3 second
)

// Cache struct for local cache
type Cache struct {
	prefix string
	cache  cache
}

type cache interface {
	Get(k string) (interface{}, bool)
	Set(k string, x interface{}, d time.Duration)
	Delete(k string)
	Flush()
}

type Opts struct {
	Prefix            string
	DefaultExpiration time.Duration
}

// New initialize local cache.
func New(o *Opts) *Cache {
	c := _localCache.New(o.DefaultExpiration, defaultCleanupInterval)
	return &Cache{
		prefix: o.Prefix,
		cache:  c,
	}
}

func (m *Cache) Get(key string) interface{} {
	value, found := m.cache.Get(key)
	if found {
		return value
	}
	return nil
}

func (m *Cache) Set(key string, value interface{}) {
	m.cache.Set(key, value, 0)
}

func (m *Cache) SetTTL(key string, value interface{}, seconds int) {
	m.cache.Set(key, value, time.Duration(seconds)*time.Second)
}

func (m *Cache) Del(keys ...string) {
	for _, key := range keys {
		m.cache.Delete(key)
	}
}

func (m *Cache) Flush() {
	m.cache.Flush()
}
