package redis

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"

	"github.com/fzerorubigd/chapar/tasks"
	"github.com/fzerorubigd/chapar/workers"
)

type redisDriver struct {
	client      *redis.Pool
	queuePrefix string

	chans map[string]chan *tasks.Task
	lock  sync.Mutex

	ctx func() (context.Context, context.CancelFunc)
}

// Options is the option type used for the create redis client or other options
type Options struct {
	client *redis.Pool
	prefix string
}

type Handler func(*Options) error

func (rd *redisDriver) pop(ctx context.Context, topic string) chan *tasks.Task {
	type duet struct {
		res interface{}
		err error
	}
	read := func() chan duet {
		c := make(chan duet)
		go func() {
			conn := rd.client.Get()
			d := duet{}
			d.res, d.err = conn.Do("BLPOP", topic, 0)
			select {
			case c <- d:
			case <-ctx.Done():
				if d.err != nil {
					return
				}
				res, err := redis.Strings(d.res, d.err)
				if err != nil {
					// TODO : log
					return
				}
				_, _ = conn.Do("LPUSH", topic, res[1])
			}
		}()
		return c
	}
	task := make(chan *tasks.Task)
	go func() {
		for {
			select {
			case msg := <-read():
				res, err := redis.Strings(msg.res, msg.err)
				if err != nil {
					// TODO : log
					continue
				}
				if len(res) == 2 && res[0] == topic {
					tsk := tasks.Task{}
					err := json.Unmarshal([]byte(res[1]), &tsk)
					if err == nil {
						// TODO : log
						select {
						case task <- &tsk:
							continue
						case <-ctx.Done():
							// OK, context canceled return the job back to the redis
							// TODO: err check? log?
							_, _ = rd.client.Get().Do("LPUSH", topic, res[1])
							return
						}
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return task
}

func (rd *redisDriver) Jobs(queue string) chan *tasks.Task {
	rd.lock.Lock()
	defer rd.lock.Unlock()

	q := rd.queuePrefix + queue
	c, ok := rd.chans[q]
	if !ok {
		ctx, _ := rd.ctx()
		c = rd.pop(ctx, q)
		rd.chans[q] = c
	}

	return c
}

func (rd *redisDriver) Sync(q string, t *tasks.Task) error {
	q = rd.queuePrefix + q
	b, err := json.Marshal(t)
	if err != nil {
		return err
	}
	_, err = redis.Int(rd.client.Get().Do("RPUSH", q, string(b)))
	return err
}

func (rd *redisDriver) Async(q string, t *tasks.Task) {
	// In redis async is not that important
	go func() {
		if err := rd.Sync(q, t); err != nil {
			// TODO: log
		}
	}()
}

// WithRedisPool is one mandatory option to set the client
func WithRedisPool(c *redis.Pool) Handler {
	return func(o *Options) error {
		if o.client != nil {
			return errors.New("client already set, only set one client option or redis option")
		}
		if c == nil {
			return errors.New("client is nil")
		}
		o.client = c
		return nil
	}
}

// WithRedisOptions try to create redis client based on options
func WithRedisOptions(network, address string, options ...redis.DialOption) Handler {
	return func(o *Options) error {
		if o.client != nil {
			return errors.New("client already set, only set one client option or redis option")
		}
		o.client = &redis.Pool{
			Dial: func() (redis.Conn, error) {
				return redis.Dial(network, address, options...)
			},
			TestOnBorrow: func(c redis.Conn, _ time.Time) error {
				_, err := c.Do("PING")
				return err
			},
			MaxIdle: 1,
		}
		return nil
	}
}

// WithQueuePrefix add prefix for queue name key
func WithQueuePrefix(p string) Handler {
	return func(o *Options) error {
		o.prefix = p
		return nil
	}
}

// NewDriver return new driver with redis backend
func NewDriver(ctx context.Context, opts ...Handler) (workers.Driver, error) {
	o := &Options{}
	for i := range opts {
		if err := opts[i](o); err != nil {
			return nil, err
		}
	}

	if o.client == nil {
		return nil, errors.New("no client set in the option sue either WithRedisOptions or WithRedisPool")
	}

	return &redisDriver{
		client:      o.client,
		queuePrefix: o.prefix,
		chans:       make(map[string]chan *tasks.Task),
		ctx: func() (context.Context, context.CancelFunc) {
			return context.WithCancel(ctx)
		},
	}, nil
}
