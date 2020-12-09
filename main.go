// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"context"
	"log/syslog"
	"os"
	"time"

	"github.com/jcjones/ocsp-l2-cache/cli"
	"github.com/jcjones/ocsp-l2-cache/common"

	blog "github.com/letsencrypt/boulder/log"
)

func getLogger(identifier string) blog.Logger {
	const defaultPriority = syslog.LOG_INFO | syslog.LOG_LOCAL0
	logProto := common.GetEnvString("SyslogProto", "")
	logAddr := common.GetEnvString("SyslogAddr", "")
	syslogger, err := syslog.Dial(logProto, logAddr, defaultPriority, identifier)
	if err != nil {
		panic(err)
	}
	logger, err := blog.New(syslogger, int(syslog.LOG_DEBUG), int(syslog.LOG_DEBUG))
	if err != nil {
		panic(err)
	}
	err = blog.Set(logger)
	if err != nil {
		panic(err)
	}
	return logger
}

func main() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "no-hostname"
	}
	identifier := common.GetEnvString("ID", hostname)
	logger := getLogger(identifier)

	c := cli.New().
		WithLogger(logger).
		WithIdentifier(identifier).
		WithListenAddr(common.GetEnvString("ListenOCSP", ":8080")).
		WithHealthListenAddr(common.GetEnvString("ListenHealth", ":8081")).
		WithRedis(common.GetEnvString("RedisHost", "redis:6379"), time.Second).
		WithCacheLifespan(common.GetEnvDuration("CacheLifespan", 24*time.Hour)).
		WithConnectionDeadline(common.GetEnvDuration("ConnectionDeadline", time.Second))

	responderMap, err := common.GetEnvMap("Responders")
	if err != nil {
		logger.Errf("Fatal decoding Responders: %v", err)
	}
	for keyId, responder := range responderMap {
		c.WithUpstreamResponder(keyId, responder)
	}

	err = c.Run(context.Background())
	if err != nil {
		logger.Errf("Fatal: %v", err)
		os.Exit(42)
	}
}
