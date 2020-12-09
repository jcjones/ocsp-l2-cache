// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package cli handles parsing input data by the main function
package cli

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/jcjones/ocsp-l2-cache/fetcher"
	"github.com/jcjones/ocsp-l2-cache/repo"
	"github.com/jcjones/ocsp-l2-cache/server"
	"github.com/jcjones/ocsp-l2-cache/storage"

	blog "github.com/letsencrypt/boulder/log"
)

type Responder struct {
	issuer       storage.Issuer
	responderUrl url.URL
}

// CLI holds state for a run of the tool; use the Run method to execute it. Can
// run more than once.
type CLI struct {
	logger             blog.Logger
	identifier         string
	listenAddr         string
	healthListenAddr   string
	redisAddr          string
	redisTxTimeout     time.Duration
	deadline           time.Duration
	lifespan           time.Duration
	upstreamResponders []Responder
}

// New constructs a Command Line Interface handler. Use its methods to configure
// it, then call the Run method to get a result.
func New() *CLI {
	return &CLI{}
}

// WithUpstreamResponder sets the URL of the upstream responder to query.
func (cli *CLI) WithUpstreamResponder(issuerId string, respUrl string) *CLI {
	rurl, err := url.Parse(respUrl)
	if err != nil {
		panic(err)
	}
	issuer, err := storage.NewIssuerFromHexKeyId(issuerId)
	if err != nil {
		panic(err)
	}
	r := Responder{
		issuer:       *issuer,
		responderUrl: *rurl,
	}

	cli.upstreamResponders = append(cli.upstreamResponders, r)
	return cli
}

func (cli *CLI) WithLogger(logger blog.Logger) *CLI {
	cli.logger = logger
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

func (cli *CLI) WithHealthListenAddr(addr string) *CLI {
	cli.healthListenAddr = addr
	return cli
}

func (cli *CLI) WithRedis(addr string, txTimeout time.Duration) *CLI {
	cli.redisAddr = addr
	cli.redisTxTimeout = txTimeout
	return cli
}

func (cli *CLI) WithCacheLifespan(responseLifespan time.Duration) *CLI {
	cli.lifespan = responseLifespan
	return cli
}

func (cli *CLI) WithConnectionDeadline(deadline time.Duration) *CLI {
	cli.deadline = deadline
	return cli
}

func (cli *CLI) Check(ctx context.Context) error {
	if cli.listenAddr == "" {
		return fmt.Errorf("Must set listen address")
	}
	if len(cli.upstreamResponders) < 1 {
		return fmt.Errorf("Must set upstream URL")
	}
	if cli.redisAddr == "" || cli.redisTxTimeout == 0 {
		return fmt.Errorf("Must set Redis address and transaction timeout")
	}
	if cli.lifespan == 0 {
		return fmt.Errorf("Must set a response lifespan")
	}
	if cli.deadline == 0 {
		return fmt.Errorf("Must set a query deadline")
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

	cli.logger.Infof("Connecting to Redis cache at %s, timeout %s", cli.redisAddr, cli.redisTxTimeout)

	startCtx, cancelFunc := context.WithTimeout(ctx, time.Second)

	remoteCache, err := storage.NewRedisCache(startCtx, cli.redisAddr, cli.redisTxTimeout)
	if err != nil {
		cancelFunc()
		return err
	}
	cancelFunc()

	store := repo.NewOcspStore(cli.logger, remoteCache, cli.lifespan, time.Hour)

	for _, r := range cli.upstreamResponders {
		upstreamFetcher, err := fetcher.NewUpstreamFetcher(r.responderUrl, cli.identifier)
		if err != nil {
			return err
		}
		err = store.AddFetcherForIssuer(r.issuer, upstreamFetcher)
		if err != nil {
			return err
		}
	}

	// Health monitoring
	hc := server.NewHealthCheck(cli.logger, remoteCache)
	healthHandler := http.NewServeMux()
	healthHandler.HandleFunc("/", hc.HandleQuery)

	healthServer := &http.Server{
		Handler: healthHandler,
		Addr:    cli.healthListenAddr,
	}
	go func() {
		if err := healthServer.ListenAndServe(); err != http.ErrServerClosed {
			cli.logger.Errf("Couldn't start health server: %v", err)
		}
	}()

	// Web-client interaction
	frontEnd, err := server.NewOcspFrontEnd(cli.logger, store, cli.deadline)
	if err != nil {
		return err
	}

	ocspServer := &http.Server{
		Addr: cli.listenAddr,
	}

	http.HandleFunc("/", frontEnd.HandleQuery)
	done := make(chan bool)
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		cli.logger.Infof("Signal caught, HTTP server shutting down.")

		// We received an interrupt signal, shut down.
		_ = ocspServer.Shutdown(ctx)
		_ = healthServer.Shutdown(ctx)
		done <- true
	}()

	fmt.Println("okay!")

	cli.logger.Infof("OCSP Serving on %v, Health Serving on %v", ocspServer.Addr, healthServer.Addr)

	for _, r := range cli.upstreamResponders {
		cli.logger.Infof("Responder key ID: %s url: %s", r.issuer, r.responderUrl.String())
	}

	if err := ocspServer.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	<-done
	cli.logger.Infof("HTTP server offline.")
	return nil
}
