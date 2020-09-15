package redis

/*
	Wrappers for Redis functionalities.
	Return simple responses without error and all the functions I frequently used.
*/

import (
	"errors"
	"fmt"
	"time"

	_redis "github.com/go-redis/redis/v7"
)

var (
	// ErrEmptyConnection empty client returned
	ErrEmptyConnection = errors.New("redis: return empty connection")
	intTrue            = int64(1)
	intFalse           = int64(0)
)

type Redis struct {
	client iRedis
	prefix string
}

type iRedis interface {
	_redis.Cmdable
}

type Opts struct {
	Address    string // along with port, e.g. localhost:6379
	Password   string // default empty, no password required.
	Prefix     string // default empty, no prefix.
	MaxRetries int    // default zero, no retry at all.
}

func New(o *Opts) (*Redis, error) {
	c := _redis.NewClient(&_redis.Options{
		Addr:       o.Address,
		Password:   o.Password,
		MaxRetries: o.MaxRetries,
	})
	if c == nil {
		return nil, ErrEmptyConnection
	}

	// test connection
	if err := c.Ping().Err(); err != nil {
		return nil, err
	}

	return &Redis{
		prefix: o.Prefix,
		client: c,
	}, nil
}

func (r *Redis) pre(key string) string {
	return fmt.Sprintf("%s:%s", r.prefix, key)
}

func (r *Redis) Ping() error {
	return r.client.Ping().Err()
}

//--- Keys

func (r *Redis) Exists(key string) bool {
	return r.client.Exists(r.pre(key)).Val() == intTrue
}

func (r *Redis) Keys(pattern string) []string {
	return r.client.Keys(pattern).Val()
}

func (r *Redis) Expire(key string, seconds int64) error {
	return r.client.Expire(r.pre(key), time.Duration(seconds)).Err()
}

func (r *Redis) Persist(key string) error {
	return r.client.Persist(r.pre(key)).Err()
}

func (r *Redis) Del(key string) error {
	return r.client.Del(r.pre(key)).Err()
}

func (r *Redis) Dump(key string) string {
	return r.client.Dump(r.pre(key)).Val()
}

//--- Strings

func (r *Redis) Get(key string) string {
	return r.client.Get(r.pre(key)).Val()
}

func (r *Redis) Set(key string, value interface{}) error {
	return r.client.Set(r.pre(key), value, 0).Err()
}

func (r *Redis) Setex(key string, value interface{}, seconds int64) error {
	return r.client.Set(r.pre(key), value, time.Duration(seconds)).Err()
}

func (r *Redis) SetNX(key string, value interface{}) error {
	return r.client.SetNX(r.pre(key), value, 0).Err()
}

func (r *Redis) SetNXex(key string, value interface{}, seconds int64) error {
	return r.client.SetNX(r.pre(key), value, time.Duration(seconds)).Err()
}

func (r *Redis) Increment(key string) error {
	return r.client.Incr(r.pre(key)).Err()
}

func (r *Redis) Decrement(key string) error {
	return r.client.Decr(r.pre(key)).Err()
}

//--- Hashes

func (r *Redis) HGet(key, field string) string {
	return r.client.HGet(r.pre(key), field).Val()
}

func (r *Redis) HGetAll(key string) map[string]string {
	return r.client.HGetAll(r.pre(key)).Val()
}

func (r *Redis) HSet(key string, fieldvalue ...interface{}) error {
	return r.client.HSet(r.pre(key), fieldvalue...).Err()
}

func (r *Redis) HSetNX(key, field string, value interface{}) error {
	return r.client.HSetNX(r.pre(key), field, value).Err()
}

func (r *Redis) HVals(key string) []string {
	return r.client.HVals(r.pre(key)).Val()
}

func (r *Redis) HDel(key, field string) error {
	return r.client.HDel(r.pre(key), field).Err()
}

func (r *Redis) HExists(key, field string) bool {
	return r.client.HExists(r.pre(key), field).Val()
}

//--- Sets

func (r *Redis) SAdd(key string, values ...interface{}) error {
	return r.client.SAdd(r.pre(key), values...).Err()
}

func (r *Redis) SMembers(key string) []string {
	return r.client.SMembers(r.pre(key)).Val()
}

func (r *Redis) SRem(key string, members ...interface{}) error {
	return r.client.SRem(r.pre(key), members...).Err()
}

func (r *Redis) ZAdd(key string, values ...interface{}) error {
	score := float64(0)
	z := make([]*_redis.Z, len(values))
	for _, v := range values {
		z = append(z, &_redis.Z{
			Score:  score,
			Member: v,
		})
		score++
	}
	return r.client.ZAdd(r.pre(key), z...).Err()
}

func (r *Redis) ZMembers(key string) []string {
	return r.client.ZRange(r.pre(key), 0, -1).Val()
}

func (r *Redis) ZRem(key string, members ...interface{}) error {
	return r.client.ZRem(r.pre(key), members...).Err()
}
