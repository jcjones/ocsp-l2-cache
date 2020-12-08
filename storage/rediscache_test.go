// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package storage

import (
	"context"
	"os"
	"testing"
	"time"
)

var kRedisHost = "RedisHost"

func getRedisCache(tb testing.TB) *RedisCache {
	setting, ok := os.LookupEnv(kRedisHost)
	if !ok {
		tb.Skipf("%s is not set, unable to run %s. Skipping.", kRedisHost, tb.Name())
	}
	tb.Logf("Connecting to Redis instance at %s", setting)

	rc, err := NewRedisCache(context.TODO(), setting, time.Second)
	if err != nil {
		tb.Fatalf("Couldn't construct RedisCache: %v", err)
	}
	return rc
}

func Test_RedisInvalidHost(t *testing.T) {
	t.Parallel()
	_, err := NewRedisCache(context.TODO(), "unknown_host:999999", time.Second)
	if err == nil {
		t.Error("Should have failed to construct invalid redis cache host")
	}
}

func Test_RedisExpiration(t *testing.T) {
	ctx := context.TODO()
	t.Parallel()
	rc := getRedisCache(t)
	defer rc.client.Del(ctx, "expTest")

	err := rc.Set(ctx, "expTest", "a", time.Hour)
	if err != nil {
		t.Error(err)
	}

	if exists, err := rc.Exists(ctx, "expTest"); exists == false || err != nil {
		t.Errorf("Should exist: %v %v", exists, err)
	}

	anHourAgo := time.Now().Add(time.Hour * -1)
	if err := rc.ExpireAt(ctx, "expTest", anHourAgo); err != nil {
		t.Error(err)
	}

	if exists, err := rc.Exists(ctx, "expTest"); exists == true || err != nil {
		t.Errorf("Should not exist anymore: %v %v", exists, err)
	}

	err = rc.Set(ctx, "expTest", "b", time.Hour)
	if err != nil {
		t.Error(err)
	}

	if err := rc.ExpireAt(ctx, "expTest", time.Now().Add(time.Second)); err != nil {
		t.Error(err)
	}

	time.Sleep(2 * time.Second)

	if exists, err := rc.Exists(ctx, "expTest"); exists == true || err != nil {
		t.Errorf("Should not exist anymore: %v %v", exists, err)
	}
}

func Test_RedisSetIfNotExist(t *testing.T) {
	ctx := context.TODO()
	t.Parallel()
	rc := getRedisCache(t)

	q := "Test_RedisSetIfNotExist"
	defer rc.client.Del(ctx, q)

	v, err := rc.SetIfNotExist(ctx, q, "me", time.Minute)
	if err != nil {
		t.Error(err)
	}
	if v != "me" {
		t.Errorf("Should have worked trivially, got %s", v)
	}

	v2, err := rc.SetIfNotExist(ctx, q, "you", time.Minute)
	if err != nil {
		t.Error(err)
	}
	if v2 != "me" {
		t.Errorf("Should not have changed from me, is now %s", v2)
	}
}

func Test_RedisGetSet(t *testing.T) {
	ctx := context.TODO()
	t.Parallel()
	rc := getRedisCache(t)

	k := "Test_RedisGetSet"
	defer rc.client.Del(ctx, k)

	_, ok, err := rc.Get(ctx, k)
	if err != nil {
		t.Error(err)
	}
	if ok {
		t.Errorf("Expected no answer for %s", k)
	}

	err = rc.Set(ctx, k, "data", time.Hour)
	if err != nil {
		t.Error(err)
	}

	v, ok, err := rc.Get(ctx, k)
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Errorf("Expected to find data")
	}
	if v != "data" {
		t.Errorf("Expected data, got %s", v)
	}
}

func Test_Info(t *testing.T) {
	ctx := context.TODO()
	t.Parallel()

	rc := getRedisCache(t)
	data, err := rc.Info(ctx)
	if err != nil {
		t.Error(err)
	}
	if len(data) == 0 {
		t.Error("Expected informational output")
	}
}
