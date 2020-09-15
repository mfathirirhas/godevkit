package redis

import (
	_mock "github.com/stretchr/testify/mock"
)

type Mock struct {
	_mock.Mock
}

func (r *Mock) Ping() error {
	return r.Called().Error(0)
}

//--- Keys

func (r *Mock) Exists(key string) bool {
	return r.Called(key).Bool(0)
}

func (r *Mock) Keys(pattern string) []string {
	return r.Called(pattern).Get(0).([]string)
}

func (r *Mock) Expire(key string, seconds int64) error {
	return r.Called(key, seconds).Error(0)
}

func (r *Mock) Persist(key string) error {
	return r.Called(key).Error(0)
}

func (r *Mock) Del(key string) error {
	return r.Called(key).Error(0)
}

func (r *Mock) Dump(key string) string {
	return r.Called(key).String(0)
}

//--- Strings

func (r *Mock) Get(key string) string {
	return r.Called(key).String(0)
}

func (r *Mock) Set(key string, value interface{}) error {
	return r.Called(key, value).Error(0)
}

func (r *Mock) Setex(key string, value interface{}, seconds int64) error {
	return r.Called(key, value, seconds).Error(0)
}

func (r *Mock) SetNX(key string, value interface{}) error {
	return r.Called(key, value).Error(0)
}

func (r *Mock) SetNXex(key string, value interface{}, seconds int64) error {
	return r.Called(key, value, seconds).Error(0)
}

func (r *Mock) Increment(key string) error {
	return r.Called(key).Error(0)
}

func (r *Mock) Decrement(key string) error {
	return r.Called(key).Error(0)
}

//--- Hashes

func (r *Mock) HGet(key, field string) string {
	return r.Called(key, field).String(0)
}

func (r *Mock) HGetAll(key string) map[string]string {
	return r.Called(key).Get(0).(map[string]string)
}

func (r *Mock) HSet(key string, fieldvalue ...interface{}) error {
	return r.Called(key, fieldvalue).Error(0)
}

func (r *Mock) HSetNX(key, field string, value interface{}) error {
	return r.Called(key, field, value).Error(0)
}

func (r *Mock) HVals(key string) []string {
	return r.Called(key).Get(0).([]string)
}

func (r *Mock) HDel(key, field string) error {
	return r.Called(key, field).Error(0)
}

func (r *Mock) HExists(key, field string) bool {
	return r.Called(key, field).Bool(0)
}

//--- Sets

func (r *Mock) SAdd(key string, values ...interface{}) error {
	return r.Called(key, values).Error(0)
}

func (r *Mock) SMembers(key string) []string {
	return r.Called(key).Get(0).([]string)
}

func (r *Mock) SRem(key string, members ...interface{}) error {
	return r.Called(key, members).Error(0)
}

func (r *Mock) ZAdd(key string, values ...interface{}) error {
	return r.Called(key, values).Error(0)
}

func (r *Mock) ZMembers(key string) []string {
	return r.Called(key).Get(0).([]string)
}

func (r *Mock) ZRem(key string, members ...interface{}) error {
	return r.Called(key, members).Error(0)
}
