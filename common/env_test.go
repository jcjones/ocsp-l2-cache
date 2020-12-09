// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package common

import (
	"os"
	"testing"
	"time"
)

func TestEnvString(t *testing.T) {
	t.Parallel()

	x := GetEnvString("TestEnvString", "default")
	if x != "default" {
		t.Errorf("Expected default, got %s", x)
	}

	_ = os.Setenv("TestEnvString", "abc")

	x = GetEnvString("TestEnvString", "default")
	if x != "abc" {
		t.Errorf("Expected abc, got %s", x)
	}
}

func TestEnvDuration(t *testing.T) {
	t.Parallel()

	x := GetEnvDuration("TestEnvDuration", time.Hour)
	if x != time.Hour {
		t.Errorf("Expected default, got %v", x)
	}

	_ = os.Setenv("TestEnvDuration", "15s")

	x = GetEnvDuration("TestEnvDuration", time.Hour)
	if x != 15*time.Second {
		t.Errorf("Expected 15 seconds, got %v", x)
	}

	_ = os.Setenv("TestEnvDuration", "3600s")

	x = GetEnvDuration("TestEnvDuration", time.Hour)
	if x != time.Hour {
		t.Errorf("Expected hour, got %v", x)
	}
}

func TestEnvMap(t *testing.T) {
	t.Parallel()

	x, err := GetEnvMap("TestEnvMap")
	if len(x) != 0 {
		t.Errorf("Expected empty array")
	}
	if err == nil {
		t.Errorf("Expected error")
	}

	_ = os.Setenv("TestEnvMap", "a=;b=;")
	x, err = GetEnvMap("TestEnvMap")
	if len(x) != 0 {
		t.Errorf("Expected empty array")
	}
	if err == nil {
		t.Errorf("Expected error")
	}

	_ = os.Setenv("TestEnvMap", "a=b;b=c;=d")
	x, err = GetEnvMap("TestEnvMap")
	if len(x) != 0 {
		t.Errorf("Expected empty array")
	}
	if err == nil {
		t.Errorf("Expected error")
	}

	_ = os.Setenv("TestEnvMap", "a=b;b=c;=d")
	x, err = GetEnvMap("TestEnvMap")
	if len(x) != 0 {
		t.Errorf("Expected empty array")
	}
	if err == nil {
		t.Errorf("Expected error")
	}

	_ = os.Setenv("TestEnvMap", "")
	x, err = GetEnvMap("TestEnvMap")
	if len(x) != 0 {
		t.Errorf("Expected empty array")
	}
	if err == nil {
		t.Errorf("Expected error")
	}

	_ = os.Setenv("TestEnvMap", "a=1;2=c;d=3.456:789")
	x, err = GetEnvMap("TestEnvMap")
	if len(x) != 3 {
		t.Errorf("Expected 3 map")
	}
	if err != nil {
		t.Error(err)
	}
}
