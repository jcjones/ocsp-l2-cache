// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package cli

import (
	"context"
	"os"
	"testing"
	"time"
)

const (
	fakeIssuerKeyId = "abcdef0123456789abcdef0123456789abcdef01"
)

func TestRunWithoutArgs(t *testing.T) {
	err := New().Run(context.TODO())
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestRunWithOnlyUpstream(t *testing.T) {
	err := New().WithUpstreamResponder(fakeIssuerKeyId, "localhost/path").Run(context.TODO())
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
	err := New().WithUpstreamResponder(fakeIssuerKeyId, "localhost/path").
		WithCacheLifespan(time.Hour).
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

	err := New().WithUpstreamResponder(fakeIssuerKeyId, "localhost/path").
		WithCacheLifespan(time.Hour).
		WithIdentifier("test").
		WithRedis(setting, time.Hour).
		WithConnectionDeadline(time.Second).
		WithListenAddr(":12345").Check(context.TODO())
	if err != nil {
		t.Fatalf("Got an error: %v", err)
	}
}
