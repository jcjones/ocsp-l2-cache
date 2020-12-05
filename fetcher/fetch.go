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

	"github.com/jcjones/ocsp-l2-cache/common"
)

var (
	RelevantHeaders = []string{
		common.HeaderCacheControl,
		common.HeaderETag,
		common.HeaderLastModified,
		common.HeaderExpires,
	}
)

func getRelevantHeaders(h http.Header) map[string]string {
	headers := make(map[string]string)

	for _, k := range RelevantHeaders {
		v := h.Get(k)
		if v != "" {
			headers[k] = v
		}
	}

	return headers
}

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
	if maxGetLen < 0 {
		return nil, fmt.Errorf("Illegal URL, how did we get here?")
	}

	return &UpstreamFetcher{
		upstreamUrl,
		maxGetLen,
		identifier,
	}, nil
}

func (uf *UpstreamFetcher) setHeaders(h *http.Header) {
	h.Add("X-Ocsp-L2-Cache", uf.identifier)
}

func (uf *UpstreamFetcher) ocspPost(ctx context.Context, ocspReq []byte) ([]byte, map[string]string, error) {
	body := bytes.NewReader(ocspReq)
	req, err := http.NewRequestWithContext(ctx, "POST", uf.upstreamUrl.String(), body)
	if err != nil {
		return []byte{}, nil, err
	}

	uf.setHeaders(&req.Header)
	req.Header.Set(common.HeaderContentType, common.MimeOcspRequest)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return []byte{}, nil, fmt.Errorf(resp.Status)
	}
	contentType := resp.Header.Get(common.HeaderContentType)
	if contentType != common.MimeOcspResponse {
		return []byte{}, nil, fmt.Errorf("Unexpected content-type: %s", contentType)
	}

	headers := getRelevantHeaders(resp.Header)

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	return data, headers, err
}

func (uf *UpstreamFetcher) ocspGet(ctx context.Context, ocspReq []byte) ([]byte, map[string]string, error) {
	b64Req := base64.RawURLEncoding.EncodeToString(ocspReq)
	url := *uf.upstreamUrl
	url.Path += "/" + b64Req
	req, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	if err != nil {
		return []byte{}, nil, err
	}

	uf.setHeaders(&req.Header)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return []byte{}, nil, fmt.Errorf(resp.Status)
	}

	contentType := resp.Header.Get(common.HeaderContentType)
	if contentType != common.MimeOcspResponse {
		return []byte{}, nil, fmt.Errorf("Unexpected content-type: %s", contentType)
	}

	headers := getRelevantHeaders(resp.Header)

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	return data, headers, err
}

func (uf *UpstreamFetcher) useGetRequest(ocspReq []byte) bool {
	return base64.RawURLEncoding.EncodedLen(len(ocspReq)) <= uf.maxGetLen
}

func (uf *UpstreamFetcher) Fetch(ctx context.Context, ocspReq []byte) ([]byte, map[string]string, error) {
	if uf.useGetRequest(ocspReq) {
		return uf.ocspPost(ctx, ocspReq)
	}
	return uf.ocspGet(ctx, ocspReq)
}
