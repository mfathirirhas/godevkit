package local

import (
	_mock "github.com/stretchr/testify/mock"
)

type Mock struct {
	_mock.Mock
}

func (m *Mock) Get(key string) interface{} {
	return m.Called(key).Get(0)

}

func (m *Mock) Set(key string, value interface{}) {
	m.Called(key, value)
}

func (m *Mock) SetTTL(key string, value interface{}, seconds int) {
	m.Called(key, value, seconds)
}

func (m *Mock) Del(keys ...string) {
	m.Called(keys)
}

func (m *Mock) Flush() {
	m.Called()
}
