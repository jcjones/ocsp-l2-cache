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

	blog "github.com/letsencrypt/boulder/log"
)

func main() {
	const defaultPriority = syslog.LOG_INFO | syslog.LOG_LOCAL0
	syslogger, err := syslog.Dial("", "", defaultPriority, "test")
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

	// TODO: read from env vars
	err = cli.New().
		WithLogger(logger).
		WithUpstreamResponder("A84A6A63047DDDBAE6D139B7A64565EFF3A8ECA1", "http://ocsp.int-x3.letsencrypt.org").
		WithUpstreamResponder("C5B1AB4E4CB1CD6430937EC1849905ABE603E225", "http://ocsp.int-x4.letsencrypt.org").
		WithUpstreamResponder("142EB317B75856CBAE500940E61FAF9D8B14C2C6", "http://r3.o.lencr.org").
		WithUpstreamResponder("369D3EE0B140F6272C7CBF8D9D318AF654A64626", "http://r4.o.lencr.org").
		WithIdentifier("jcj testing").
		WithListenAddr(":8080").
		WithHealthListenAddr(":8081").
		WithRedis("192.168.99.100:6379", time.Second).
		WithCacheLifespan(24 * time.Hour).
		WithConnectionDeadline(time.Second).
		// Signals are handled in the CLI package
		Run(context.Background())

	if err != nil {
		logger.Errf("Fatal due to error %v, exiting with code 42", err)
		os.Exit(42)
	}
}
