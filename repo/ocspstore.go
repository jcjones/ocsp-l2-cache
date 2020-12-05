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
	responders map[string]fetcher.UpstreamFetcher
	cache      storage.RemoteCache
	lifespan   time.Duration
	minimumCacheLife time.Duration
}

func NewOcspStore(cache storage.RemoteCache, lifespan time.Duration, minimumCacheLife time.Duration) OcspStore {
	return OcspStore{
		make(map[string]fetcher.UpstreamFetcher),
		cache,
		lifespan,
		minimumCacheLife,
	}
}

func (c *OcspStore) AddFetcherForIssuer(issuer storage.Issuer, uf *fetcher.UpstreamFetcher) error {
	if uf == nil {
		return fmt.Errorf("Fetcher must not be nil")
	}

	c.responders[issuer.String()] = *uf
	return nil
}

func (c *OcspStore) Get(ctx context.Context, req *ocsp.Request, reqBytes []byte) ([]byte, map[string]string, error) {
	issuer := storage.NewIssuerFromRequest(req)
	uf, ok := c.responders[issuer.String()]
	if !ok {
		return nil, nil, UnknownIssuerError
	}

	serial, err := storage.NewSerialFromBigInt(req.SerialNumber)
	if err != nil {
		return nil, nil, err
	}

	cacheRsp, found, err := c.cache.Get(ctx, serial.BinaryString())
	if err != nil {
		return nil, nil, err
	}

	if found {
		cr, err := NewCompressedResponseFromBinaryString(cacheRsp, serial)
		if err != nil {
			return nil, nil, err
		}

		log.Printf("issuer %s serial %s hit", issuer.String(), serial.String())
		return cr.RawResp, cr.Headers(), nil
	}

	log.Printf("issuer %s serial %s miss", issuer.String(), serial.String())

	rspBytes, headers, err := uf.Fetch(ctx, reqBytes)
	if err != nil {
		log.Printf("Fetch error: %v", err)
		return nil, nil, UpstreamError
	}

	// Don't verify here, use nil as issuer
	resp, err := ocsp.ParseResponse(rspBytes, nil)
	if err != nil {
		log.Printf("Parse of upstream response error: %v", err)
		return nil, nil, UpstreamError
	}

	cacheEndTime := resp.ThisUpdate.Add(c.lifespan)
	remainingLife := time.Until(cacheEndTime)
	if remainingLife < c.minimumCacheLife {
		remainingLife = c.minimumCacheLife
	}

	cr, err := NewCompressedResponseFromRawResponseAndHeaders(rspBytes, headers)
	if err != nil {
		return nil, nil, err
	}

	encoded, err := cr.BinaryString()
	if err != nil {
		return nil, nil, err
	}

	err = c.cache.Set(ctx, serial.BinaryString(), encoded, remainingLife)
	if err != nil {
		return nil, nil, err
	}

	return rspBytes, headers, nil
}
