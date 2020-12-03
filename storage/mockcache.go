// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package storage

import (
	"context"
	"path/filepath" // used for glob-like matching in Keys
	"time"
)

type MockRemoteCache struct {
	Data        map[string]string
	Expirations map[string]time.Time
	Duplicate   int
}

func NewMockRemoteCache() *MockRemoteCache {
	return &MockRemoteCache{
		Data:        make(map[string]string),
		Expirations: make(map[string]time.Time),
		Duplicate:   0,
	}
}

func (ec *MockRemoteCache) CleanupExpiry() {
	now := time.Now()
	for key, timestamp := range ec.Expirations {
		if timestamp.Before(now) {
			delete(ec.Data, key)
			delete(ec.Expirations, key)
		}
	}
}

func (ec *MockRemoteCache) Exists(ctx context.Context, key string) (bool, error) {
	ec.CleanupExpiry()
	_, ok := ec.Data[key]
	return ok, nil
}

func (ec *MockRemoteCache) ExpireAt(ctx context.Context, key string, expTime time.Time) error {
	ec.Expirations[key] = expTime
	return nil
}

func (ec *MockRemoteCache) KeysToChan(ctx context.Context, pattern string, c chan<- string) error {
	defer close(c)

	for key := range ec.Data {
		matched, err := filepath.Match(pattern, key)
		if err != nil {
			return err
		}
		if matched {
			c <- key
		}
	}

	return nil
}

func (ec *MockRemoteCache) SetIfNotExist(ctx context.Context, key string, v string, life time.Duration) (string, error) {
	val, ok := ec.Data[key]
	if ok {
		return val, nil
	}
	ec.Data[key] = v
	err := ec.ExpireAt(ctx, key, time.Now().Add(life))
	return v, err
}

func (ec *MockRemoteCache) Set(ctx context.Context, k string, v string, life time.Duration) error {
	ec.Data[k] = v
	return ec.ExpireAt(ctx, k, time.Now().Add(life))
}

func (ec *MockRemoteCache) Get(ctx context.Context, k string) (string, bool, error) {
	ec.CleanupExpiry()
	v, ok := ec.Data[k]
	return v, ok, nil
}
