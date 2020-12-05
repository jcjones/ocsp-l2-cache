// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jcjones/ocsp-l2-cache/cli"
)

// Accept CLI arguments and pass them to the internal methods, to permit testing
func main() {
	// TODO: pull in an argparse
	err := cli.New().
		WithUpstreamResponderURL("http://ocsp.int-x3.letsencrypt.org").
		WithIdentifier("jcj testing").
		WithListenAddr(":9020").
		WithRedis("192.168.99.100:6379", time.Second).
		WithLifespan(4 * 24 * time.Hour).
		// Signals are handled in the CLI package
		Run(context.Background())

	if err != nil {
		log.Printf("Fatal due to error %v, exiting with code 42", err)
		os.Exit(42)
	}
}
