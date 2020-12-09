// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package common

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func GetEnvString(name string, def string) string {
	setting, ok := os.LookupEnv(name)
	if !ok {
		return def
	}
	return setting
}

func GetEnvDuration(name string, def time.Duration) time.Duration {
	setting, ok := os.LookupEnv(name)
	if !ok {
		return def
	}
	dur, err := time.ParseDuration(setting)
	if err != nil {
		return def
	}
	return dur
}

func GetEnvMap(name string) (map[string]string, error) {
	var nilmap map[string]string

	setting, ok := os.LookupEnv(name)
	if !ok {
		return nilmap, fmt.Errorf("Variable not set")
	}

	data := make(map[string]string)

	segments := strings.Split(setting, ";")
	for _, part := range segments {
		parts := strings.Split(part, "=")
		if len(parts) != 2 {
			return nilmap, fmt.Errorf("Segment %s has %d parts", part, len(parts))
		}
		if len(parts[0]) == 0 || len(parts[1]) == 0 {
			return nilmap, fmt.Errorf("Empty entry: %s", part)
		}
		data[parts[0]] = parts[1]
	}

	return data, nil
}
