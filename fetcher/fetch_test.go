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
)

func TestFetchNilUrl(t *testing.T) {
	_, err := NewUpstreamFetcher(nil, "TestFetchNilUrl")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestFetch404(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	url, _ := url.Parse(ts.URL)

	f, err := NewUpstreamFetcher(url, "TestFetch404")
	if err != nil {
		t.Error(err)
	}

	v, err := f.ocspGet(context.TODO(), []byte{})
	if err == nil {
		t.Error("Expected error")
	}
	if len(v) != 0 {
		t.Error("Expected no response")
	}

	v, err = f.ocspPost(context.TODO(), []byte{})
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

	f, err := NewUpstreamFetcher(url, "TestFetchBogusResponse")
	if err != nil {
		t.Error(err)
	}

	v, err := f.ocspGet(context.TODO(), []byte{})
	if err == nil {
		t.Error("Expected error")
	}
	if len(v) != 0 {
		t.Error("Expected no response")
	}

	v, err = f.ocspPost(context.TODO(), []byte{})
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

	fShort, err := NewUpstreamFetcher(shortUrl, "short")
	if err != nil {
		t.Error(err)
	}
	if !fShort.useGetRequest(make([]byte, 150)) {
		t.Error("Short URLs should use GET with 150-byte OCSP requests")
	}
	if fShort.useGetRequest(make([]byte, 400)) {
		t.Error("Short URLs should use POST with 400-byte OCSP requests")
	}

	fLong, err := NewUpstreamFetcher(longUrl, "long")
	if err != nil {
		t.Error(err)
	}
	if fLong.useGetRequest(make([]byte, 150)) {
		t.Error("Long URLs should use POST with 150-byte OCSP requests")
	}
	if fLong.useGetRequest(make([]byte, 400)) {
		t.Error("Long URLs should use POST with 400-byte OCSP requests")
	}

	_, err = NewUpstreamFetcher(brokenlyLongUrl, "broken")
	if err == nil {
		t.Error("Don't allow brokenly-long URLs")
	}
}