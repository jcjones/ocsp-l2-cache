// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package repo

import (
	"reflect"
	"testing"

	"github.com/jcjones/ocsp-l2-cache/common"
	"github.com/jcjones/ocsp-l2-cache/storage"
)

func TestInvalidBinaryString(t *testing.T) {
	_, err := NewCompressedResponseFromBinaryString("this is binary, right?", storage.NewSerialFromHex("de4d"))
	if err == nil {
		t.Error("Expected an error!")
	}
}

func TestRequiredHeaders(t *testing.T) {
	h := make(map[string]string)

	binString := []byte("hey")

	_, err := NewCompressedResponseFromRawResponseAndHeaders(binString, h); if err == nil {
		t.Error("Expected an error!")	
	}

	h[common.HeaderETag] = "etag"
	_, err = NewCompressedResponseFromRawResponseAndHeaders(binString, h); if err == nil {
		t.Error("Expected an error!")	
	}

	h[common.HeaderExpires] = "expires"
	_, err = NewCompressedResponseFromRawResponseAndHeaders(binString, h); if err == nil {
		t.Error("Expected an error!")	
	}

	h[common.HeaderLastModified] = "modified"
	_, err = NewCompressedResponseFromRawResponseAndHeaders(binString, h); if err == nil {
		t.Error("Expected an error!")	
	}

	h["irrelevant-header"] = "still shouldn't work"
	_, err = NewCompressedResponseFromRawResponseAndHeaders(binString, h); if err == nil {
		t.Error("Expected an error!")	
	}

	h[common.HeaderCacheControl] = "control"
	cr, err := NewCompressedResponseFromRawResponseAndHeaders(binString, h); if err != nil {
		t.Error(err)
	}

	delete(h, "irrelevant-header")
	if !reflect.DeepEqual(h, cr.Headers()) {
		t.Errorf("Expected the headers to match, got %+v expected %+v", cr.Headers(), h)
	}
}

func TestRoundtrip(t *testing.T) {
	h := make(map[string]string)
	h[common.HeaderETag] = "etag"
	h[common.HeaderExpires] = "expires"
	h[common.HeaderLastModified] = "modified"
	h[common.HeaderCacheControl] = "control"

	binString := []byte("hey")
	cr, err := NewCompressedResponseFromRawResponseAndHeaders(binString, h); if err != nil {
		t.Error(err)
	}

	encoded, err := cr.BinaryString(); if err != nil {
		t.Error(err)
	}

	cr2, err := NewCompressedResponseFromBinaryString(encoded, storage.NewSerialFromHex("de4d")); if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(*cr, cr2) {
		t.Errorf("Expected equality between %+v and %+v", cr, cr2)
	}	
}