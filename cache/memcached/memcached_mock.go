package memcached

import (
	_mock "github.com/stretchr/testify/mock"
)

// Mock mock for invoker
type Mock struct {
	_mock.Mock
}

func (m *Mock) Ping() error {
	r := m.Called()

	return r.Error(0)
}

func (m *Mock) Get(key string) ([]byte, error) {
	r := m.Called(key)

	return r.Get(0).([]byte), r.Error(1)
}

func (m *Mock) Set(key string, value []byte) error {
	r := m.Called(key, value)

	return r.Error(0)
}

func (m *Mock) SetTTL(key string, value []byte, seconds int) error {
	r := m.Called(key, value, seconds)

	return r.Error(0)
}

func (m *Mock) Del(keys ...string) error {
	r := m.Called(keys)

	return r.Error(0)
}

func (m *Mock) Add(key string, value []byte) error {
	return m.Called(key, value).Error(0)
}

func (m *Mock) AddTTL(key string, value []byte, seconds int) error {
	return m.Called(key, value, seconds).Error(0)
}

func (m *Mock) CompareAndSwap(key string, value []byte, seconds int) error {
	return m.Called(key, value, seconds).Error(0)
}

func (m *Mock) Decrement(key string, delta uint64) (newValue uint64, err error) {
	r := m.Called(key, delta)
	return r.Get(0).(uint64), r.Error(1)
}

func (m *Mock) DeleteAll() error {
	return m.Called().Error(0)
}

func (m *Mock) FlushAll() error {
	return m.Called().Error(0)
}

func (m *Mock) GetMulti(keys []string) (map[string][]byte, error) {
	r := m.Called(keys)
	return r.Get(0).(map[string][]byte), r.Error(1)
}

func (m *Mock) Increment(key string, delta uint64) (newValue uint64, err error) {
	r := m.Called(key, delta)
	return r.Get(0).(uint64), r.Error(1)
}

func (m *Mock) Replace(key string, value []byte) error {
	return m.Called(key, value).Error(0)
}

func (m *Mock) Touch(key string, seconds int32) (err error) {
	return m.Called(key, seconds).Error(0)
}
