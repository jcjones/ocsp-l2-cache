// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jcjones/ocsp-l2-cache/common"
)

func TestFetch404(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	url, _ := url.Parse(ts.URL)

	f, err := NewUpstreamFetcher(*url, "TestFetch404")
	if err != nil {
		t.Error(err)
	}

	v, _, err := f.ocspGet(context.TODO(), []byte{})
	if err == nil {
		t.Error("Expected error")
	}
	if len(v) != 0 {
		t.Error("Expected no response")
	}

	v, _, err = f.ocspPost(context.TODO(), []byte{})
	if err == nil {
		t.Error("Expected error")
	}
	if len(v) != 0 {
		t.Error("Expected no response")
	}
}

func TestFetchNoContentType(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello")
	}))
	defer ts.Close()

	url, _ := url.Parse(ts.URL)

	f, err := NewUpstreamFetcher(*url, "TestFetchBogusResponse")
	if err != nil {
		t.Error(err)
	}

	v, _, err := f.ocspGet(context.TODO(), []byte{})
	if err == nil {
		t.Error("Expected error")
	}
	if len(v) != 0 {
		t.Error("Expected no response")
	}

	v, _, err = f.ocspPost(context.TODO(), []byte{})
	if err == nil {
		t.Error("Expected error")
	}
	if len(v) != 0 {
		t.Error("Expected no response")
	}
}

func TestUseGetRequest(t *testing.T) {
	shortUrl, _ := url.Parse("http://example.com/")
	longUrl, _ := url.Parse("http://example.com/" + strings.Repeat("a", 253))
	brokenlyLongUrl, _ := url.Parse("http://example.com/" + strings.Repeat("a", 254))

	fShort, err := NewUpstreamFetcher(*shortUrl, "short")
	if err != nil {
		t.Error(err)
	}
	if !fShort.useGetRequest(make([]byte, 150)) {
		t.Error("Short URLs should use GET with 150-byte OCSP requests")
	}
	if fShort.useGetRequest(make([]byte, 400)) {
		t.Error("Short URLs should use POST with 400-byte OCSP requests")
	}

	fLong, err := NewUpstreamFetcher(*longUrl, "long")
	if err != nil {
		t.Error(err)
	}
	if fLong.useGetRequest(make([]byte, 150)) {
		t.Error("Long URLs should use POST with 150-byte OCSP requests")
	}
	if fLong.useGetRequest(make([]byte, 400)) {
		t.Error("Long URLs should use POST with 400-byte OCSP requests")
	}

	_, err = NewUpstreamFetcher(*brokenlyLongUrl, "broken")
	if err == nil {
		t.Error("Don't allow brokenly-long URLs")
	}
}

func checkHeader(t *testing.T, h map[string]string, k string) {
	_, ok := h[k]
	if !ok {
		t.Errorf("Expected %s got %+v", k, h)
	}
}

func TestFetchRelevantHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("irrelevant", "stuff")
		w.Header().Add(common.HeaderCacheControl, "ok")
		w.Header().Add(common.HeaderETag, "ok")
		w.Header().Add(common.HeaderLastModified, "ok")
		w.Header().Add(common.HeaderExpires, "ok")
		w.Header().Add(common.HeaderContentType, common.MimeOcspResponse)
		fmt.Fprintln(w, "bogus data")
	}))
	defer ts.Close()

	url, _ := url.Parse(ts.URL)

	f, err := NewUpstreamFetcher(*url, "TestFetchRelevantHeaders")
	if err != nil {
		t.Error(err)
	}

	_, h, err := f.ocspGet(context.TODO(), []byte{})
	if err != nil {
		t.Error(err)
	}
	if len(h) != 4 {
		t.Errorf("Expected 4 headers, got %d %+v", len(h), h)
	}

	checkHeader(t, h, "Cache-Control")
	checkHeader(t, h, "ETag")
	checkHeader(t, h, "Last-Modified")
	checkHeader(t, h, "Expires")
}
