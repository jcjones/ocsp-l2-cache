// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package repo gathers and maintains OCSP responses
package repo

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jcjones/ocsp-l2-cache/fetcher"
	"github.com/jcjones/ocsp-l2-cache/storage"
	"golang.org/x/crypto/ocsp"
)

type OcspStore struct {
	uf       *fetcher.UpstreamFetcher
	cache    storage.RemoteCache
	lifespan time.Duration
}

func NewOcspStore(uf *fetcher.UpstreamFetcher, cache storage.RemoteCache, lifespan time.Duration) (*OcspStore, error) {
	if uf == nil {
		return nil, fmt.Errorf("Fetcher must not be nil")
	}

	return &OcspStore{
		uf,
		cache,
		lifespan,
	}, nil
}

func (c *OcspStore) Get(ctx context.Context, req *ocsp.Request, reqBytes []byte) ([]byte, error) {
	serial, err := storage.NewSerialFromBigInt(req.SerialNumber)
	if err != nil {
		return nil, err
	}

	cacheRsp, found, err := c.cache.Get(ctx, serial.BinaryString())
	if err != nil {
		return nil, err
	}

	if found {
		return []byte(cacheRsp), nil
	}

	rspBytes, err := c.uf.Fetch(ctx, reqBytes)
	if err != nil {
		log.Printf("Fetch error: %v", err)
		return nil, UpstreamError
	}

	// Don't verify here, use nil as issuer
	resp, err := ocsp.ParseResponse(rspBytes, nil)
	if err != nil {
		log.Printf("Parse of upstream response error: %v", err)
		return nil, UpstreamError
	}

	cacheEndTime := resp.ThisUpdate.Add(c.lifespan)
	remainingLife := time.Until(cacheEndTime)

	err = c.cache.Set(ctx, serial.BinaryString(), string(rspBytes), remainingLife)
	if err != nil {
		return nil, err
	}

	return rspBytes, nil
}
