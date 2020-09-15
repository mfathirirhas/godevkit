package memcached

import (
	"errors"
	"fmt"

	_memcached "github.com/bradfitz/gomemcache/memcache"
)

var (
	ErrSetServers = errors.New("memcached: failed connecting to servers")
	ErrInitFailed = errors.New("memcached: failed initiating memcached")
	ErrEmptyValue = errors.New("memcached: values cannot be empty")
)

type Memcached struct {
	prefix   string
	memcache memcache
}

type memcache interface {
	Ping() error
	Get(key string) (item *_memcached.Item, err error)
	Set(item *_memcached.Item) error
	Delete(key string) error

	Add(item *_memcached.Item) error
	CompareAndSwap(item *_memcached.Item) error
	Decrement(key string, delta uint64) (newValue uint64, err error)
	DeleteAll() error
	FlushAll() error
	GetMulti(keys []string) (map[string]*_memcached.Item, error)
	Increment(key string, delta uint64) (newValue uint64, err error)
	Replace(item *_memcached.Item) error
	Touch(key string, seconds int32) (err error)
}

// Deps dependencies to initiate memcache
type Opts struct {
	Prefix string
	Addrs  []string // collection of memcache servers address with port. I.e. "10.0.0.1:11211"
}

// Init initiate memcache instance
func New(o *Opts) (*Memcached, error) {
	ss := &_memcached.ServerList{}
	err := ss.SetServers(o.Addrs...)
	if err != nil {
		return nil, ErrSetServers
	}
	m := _memcached.NewFromSelector(ss)
	if m == nil {
		return nil, ErrInitFailed
	}

	memcached := &Memcached{
		prefix:   o.Prefix,
		memcache: m,
	}

	return memcached, nil
}

func (m *Memcached) pre(key string) string {
	return fmt.Sprintf("%s:%s", m.prefix, key)
}

func (m *Memcached) Ping() error {
	return m.memcache.Ping()
}

func (m *Memcached) Get(key string) ([]byte, error) {
	r, err := m.memcache.Get(m.pre(key))
	if err != nil {
		return nil, err
	}
	return r.Value, err
}

func (m *Memcached) Set(key string, value []byte) error {
	if value == nil {
		return ErrEmptyValue
	}
	return m.memcache.Set(&_memcached.Item{
		Key:   m.pre(key),
		Value: value,
	})
}

func (m *Memcached) SetTTL(key string, value []byte, seconds int) error {
	if value == nil {
		return ErrEmptyValue
	}
	return m.memcache.Set(&_memcached.Item{
		Key:        m.pre(key),
		Value:      value,
		Expiration: int32(seconds),
	})
}

func (m *Memcached) Del(keys ...string) error {
	for _, key := range keys {
		err := m.memcache.Delete(m.pre(key))
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Memcached) Add(key string, value []byte) error {
	if value == nil {
		return ErrEmptyValue
	}

	return m.memcache.Add(&_memcached.Item{
		Key:   key,
		Value: value,
	})
}

func (m *Memcached) AddTTL(key string, value []byte, seconds int) error {
	if value == nil {
		return ErrEmptyValue
	}

	return m.memcache.Add(&_memcached.Item{
		Key:        key,
		Value:      value,
		Expiration: int32(seconds),
	})
}

func (m *Memcached) CompareAndSwap(key string, value []byte, seconds int) error {
	if value == nil {
		return ErrEmptyValue
	}

	return m.memcache.Add(&_memcached.Item{
		Key:   key,
		Value: value,
	})
}

func (m *Memcached) Decrement(key string, delta uint64) (newValue uint64, err error) {
	return m.memcache.Decrement(key, delta)
}

func (m *Memcached) DeleteAll() error {
	return m.memcache.DeleteAll()
}

func (m *Memcached) FlushAll() error {
	return m.memcache.FlushAll()
}

func (m *Memcached) GetMulti(keys []string) (map[string][]byte, error) {
	ret := make(map[string][]byte)
	r, err := m.memcache.GetMulti(keys)
	if err != nil {
		return nil, err
	}
	for k, v := range r {
		ret[k] = v.Value
	}
	return ret, nil
}

func (m *Memcached) Increment(key string, delta uint64) (newValue uint64, err error) {
	return m.memcache.Increment(key, delta)
}

func (m *Memcached) Replace(key string, value []byte) error {
	if value == nil {
		return ErrEmptyValue
	}

	return m.memcache.Replace(&_memcached.Item{
		Key:   key,
		Value: value,
	})
}

func (m *Memcached) Touch(key string, seconds int32) (err error) {
	return m.memcache.Touch(key, seconds)
}
