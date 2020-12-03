// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRunWithoutArgs(t *testing.T) {
	err := New().Run(context.TODO())
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestRunWithOnlyUpstream(t *testing.T) {
	err := New().WithUpstreamResponderURL("localhost/path").Run(context.TODO())
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestRunWithOnlyListen(t *testing.T) {
	err := New().WithListenAddr(":12345").Run(context.TODO())
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestRunWithoutRedis(t *testing.T) {
	err := New().WithUpstreamResponderURL("localhost/path").
		WithLifespan(time.Hour).
		WithListenAddr(":12345").Run(context.TODO())
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestCheck(t *testing.T) {
	setting, ok := os.LookupEnv("RedisHost")
	if !ok {
		t.Skipf("RedisHost is not set, unable to run %s. Skipping.", t.Name())
		return
	}

	err := New().WithUpstreamResponderURL("localhost/path").
		WithLifespan(time.Hour).
		WithIdentifier("test").
		WithRedis(setting, time.Hour).
		WithListenAddr(":12345").Check(context.TODO())
	if err != nil {
		t.Fatalf("Got an error: %v", err)
	}
}
