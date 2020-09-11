package redis

import (
	"errors"

	_redis "github.com/go-redis/redis/v7"
)

var (
	// ErrEmptyConnection empty client returned
	ErrEmptyConnection = errors.New("redis: return empty connection")
)

type Redis struct {
	client _redis.Cmdable
	prefix string
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

func (r *Redis) Ping() error {
	return r.client.Ping().Err()
}
