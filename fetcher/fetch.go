// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package fetch loads OCSP responses from another OCSP responder
package fetcher

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type UpstreamFetcher struct {
	upstreamUrl *url.URL
	maxGetLen   int
	identifier  string
}

func NewUpstreamFetcher(upstreamUrl *url.URL, identifier string) (*UpstreamFetcher, error) {
	if upstreamUrl == nil {
		return nil, fmt.Errorf("Upstream URL must not be nil")
	}

	maxGetLen := 254 - len(upstreamUrl.Path)

	return &UpstreamFetcher{
		upstreamUrl,
		maxGetLen,
		identifier,
	}, nil
}

func (uf *UpstreamFetcher) setHeaders(h *http.Header) {
	h.Add("X-Ocsp-L2-Cache", uf.identifier)
}

func (uf *UpstreamFetcher) ocspPost(ctx context.Context, ocspReq []byte) ([]byte, error) {
	body := bytes.NewReader(ocspReq)
	req, err := http.NewRequestWithContext(ctx, "POST", uf.upstreamUrl.String(), body)
	if err != nil {
		return []byte{}, err
	}

	uf.setHeaders(&req.Header)
	req.Header.Set("Content-Type", "application/ocsp-request")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf(resp.Status)
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (uf *UpstreamFetcher) ocspGet(ctx context.Context, ocspReq []byte) ([]byte, error) {
	b64Req := base64.URLEncoding.EncodeToString(ocspReq)
	url := *uf.upstreamUrl
	url.Path += "/" + b64Req
	fmt.Println(url.String())
	req, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	if err != nil {
		return []byte{}, err
	}

	uf.setHeaders(&req.Header)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf(resp.Status)
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (uf *UpstreamFetcher) Fetch(ctx context.Context, ocspReq []byte) ([]byte, error) {
	if true || base64.URLEncoding.EncodedLen(len(ocspReq)) > uf.maxGetLen {
		return uf.ocspPost(ctx, ocspReq)
	}
	return uf.ocspGet(ctx, ocspReq)
}
