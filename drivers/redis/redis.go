package redis

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"

	"github.com/fzerorubigd/chapar/tasks"
	"github.com/fzerorubigd/chapar/workers"
)

type redisDriver struct {
	client      *redis.Client
	queuePrefix string

	chans map[string]chan *tasks.Task
	lock  sync.Mutex

	ctx func() (context.Context, context.CancelFunc)
}

// Options is the option type used for the create redis client or other options
type Options struct {
	client *redis.Client
	prefix string
}

type Handler func(*Options) error

func (rd *redisDriver) pop(ctx context.Context, topic string) chan *tasks.Task {
	read := func() chan *redis.StringSliceCmd {
		c := make(chan *redis.StringSliceCmd)
		go func() {
			c <- rd.client.BLPop(time.Second, topic)
		}()
		return c
	}
	task := make(chan *tasks.Task)
	go func() {
		for {
			select {
			case msg := <-read():
				res, err := msg.Result()
				if err == redis.Nil {
					// ok, continue
					continue
				}
				if err != nil {
					// Log!
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
							rd.client.LPush(topic, res[1])
							return
						}
					}
				}
				// TODO : log
				// TODO: after context done, there is a chance to get a job from the redis, and we might miss that
				// case <-ctx.Done():
				// 	return
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
	i := rd.client.RPush(q, string(b))
	return i.Err()
}

func (rd *redisDriver) Async(q string, t *tasks.Task) {
	// In redis async is not that important
	// go func() {
	if err := rd.Sync(q, t); err != nil {
		// TODO: log
	}
	// }()
}

// WithRedisClient is one mandatory option to set the client
func WithRedisClient(c *redis.Client) Handler {
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
func WithRedisOptions(opt *redis.Options) Handler {
	return func(o *Options) error {
		if o.client != nil {
			return errors.New("client already set, only set one client option or redis option")
		}
		o.client = redis.NewClient(opt)
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
		return nil, errors.New("no client set in the option sue either WithRedisOptions or WithRedisClient")
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
