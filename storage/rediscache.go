// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package storage

import (
	"context"
	"time"

	"github.com/armon/go-metrics"
	"github.com/go-redis/redis/v8"
)

const NO_EXPIRATION time.Duration = 0

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(ctx context.Context, addr string, cacheTxTimeout time.Duration) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:            addr,
		MaxRetries:      10,
		MaxRetryBackoff: 5 * time.Second,
		ReadTimeout:     cacheTxTimeout,
		WriteTimeout:    cacheTxTimeout,
	})

	statusr := rdb.Ping(ctx)
	if statusr.Err() != nil {
		return nil, statusr.Err()
	}

	return &RedisCache{rdb}, nil
}

func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	defer metrics.MeasureSince([]string{"Exists"}, time.Now())
	ir := rc.client.Exists(ctx, key)
	count, err := ir.Result()
	return count == 1, err
}

func (rc *RedisCache) ExpireAt(ctx context.Context, key string, aExpTime time.Time) error {
	defer metrics.MeasureSince([]string{"ExpireAt"}, time.Now())
	br := rc.client.ExpireAt(ctx, key, aExpTime)
	return br.Err()
}

func (rc *RedisCache) KeysToChan(ctx context.Context, pattern string, c chan<- string) error {
	defer close(c)
	defer metrics.MeasureSince([]string{"KeysToChan"}, time.Now())
	scanres := rc.client.Scan(ctx, 0, pattern, 0)
	err := scanres.Err()
	if err != nil {
		return err
	}

	iter := scanres.Iterator()

	for iter.Next(ctx) {
		c <- iter.Val()
	}

	return iter.Err()
}

func (rc *RedisCache) SetIfNotExist(ctx context.Context, k string, v string, life time.Duration) (string, error) {
	br := rc.client.SetNX(ctx, k, v, life)
	if br.Err() != nil {
		return "", br.Err()
	}
	sr := rc.client.Get(ctx, k)
	return sr.Result()
}

func (rc *RedisCache) Set(ctx context.Context, k string, v string, life time.Duration) error {
	br := rc.client.SetEX(ctx, k, v, life)
	return br.Err()
}

func (rc *RedisCache) Get(ctx context.Context, k string) (string, bool, error) {
	sr := rc.client.Get(ctx, k)
	v, err := sr.Result()
	if err == redis.Nil {
		return v, false, nil
	}
	return v, true, err
}

func (rc *RedisCache) Info(ctx context.Context) (string, error) {
	sr := rc.client.Info(ctx)
	return sr.Result()
}