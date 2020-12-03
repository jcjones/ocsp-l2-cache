// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package cli handles parsing input data by the main function
package cli

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jcjones/ocsp-l2-cache/fetcher"
	"github.com/jcjones/ocsp-l2-cache/repo"
	"github.com/jcjones/ocsp-l2-cache/server"
	"github.com/jcjones/ocsp-l2-cache/storage"
)

// CLI holds state for a run of the tool; use the Run method to execute it. Can
// run more than once.
type CLI struct {
	identifier     string
	listenAddr     string
	redisAddr      string
	redisTxTimeout time.Duration
	lifespan       time.Duration
	upstreamUrl    *url.URL
}

// New constructs a Command Line Interface handler. Use its methods to configure
// it, then call the Run method to get a result.
func New() *CLI {
	return &CLI{}
}

// WithUpstreamResponderURL sets the URL of the upstream responder to query.
func (cli *CLI) WithUpstreamResponderURL(respUrl string) *CLI {
	u, err := url.Parse(respUrl)
	if err != nil {
		panic(err)
	}
	cli.upstreamUrl = u
	return cli
}

func (cli *CLI) WithIdentifier(identifier string) *CLI {
	cli.identifier = identifier
	return cli
}

// WithListenAddr sets the address:port on which to listen for queries
func (cli *CLI) WithListenAddr(addr string) *CLI {
	cli.listenAddr = addr
	return cli
}

func (cli *CLI) WithRedis(addr string, txTimeout time.Duration) *CLI {
	cli.redisAddr = addr
	cli.redisTxTimeout = txTimeout
	return cli
}

func (cli *CLI) WithLifespan(responseLifespan time.Duration) *CLI {
	cli.lifespan = responseLifespan
	return cli
}

func (cli *CLI) Check(ctx context.Context) error {
	if cli.listenAddr == "" {
		return fmt.Errorf("Must set listen address")
	}
	if cli.upstreamUrl == nil {
		return fmt.Errorf("Must set upstream URL")
	}
	if cli.redisAddr == "" || cli.redisTxTimeout == 0 {
		return fmt.Errorf("Must set Redis address and transaction timeout")
	}
	if cli.lifespan == 0 {
		return fmt.Errorf("Must set a response lifespan")
	}
	if cli.identifier == "" {
		return fmt.Errorf("Must set an identifier")
	}
	return nil
}

// Run the command, obeying the context.
func (cli *CLI) Run(ctx context.Context) error {
	err := cli.Check(ctx)
	if err != nil {
		return err
	}

	upstreamFetcher, err := fetcher.NewUpstreamFetcher(cli.upstreamUrl, cli.identifier)
	if err != nil {
		return err
	}
	remoteCache, err := storage.NewRedisCache(ctx, cli.redisAddr, cli.redisTxTimeout)
	if err != nil {
		return err
	}
	store, err := repo.NewOcspStore(upstreamFetcher, remoteCache, cli.lifespan)
	if err != nil {
		return err
	}
	frontEnd, err := server.NewOcspFrontEnd(cli.listenAddr, store)
	if err != nil {
		return err
	}
	return frontEnd.ListenAndServe()
}
