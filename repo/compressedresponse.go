package repo

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/jcjones/ocsp-l2-cache/common"
	"github.com/jcjones/ocsp-l2-cache/storage"
)

type CompressedResponse struct {
	RawResp                                   []byte
	CacheControl, ETag, LastModified, Expires string
}

func NewCompressedResponseFromBinaryString(s string, serial storage.Serial) (CompressedResponse, error) {
	var cr CompressedResponse
	buf := bytes.NewBuffer([]byte(s))
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&cr)
	return cr, err
}

func NewCompressedResponseFromRawResponseAndHeaders(RawResp []byte, headers map[string]string) (*CompressedResponse, error) {
	CacheControl, ok := headers[common.HeaderCacheControl]
	if !ok {
		return nil, fmt.Errorf("Cache-Control header not provided")
	}
	ETag, ok := headers[common.HeaderETag]
	if !ok {
		return nil, fmt.Errorf("ETag header not provided")
	}
	LastModified, ok := headers[common.HeaderLastModified]
	if !ok {
		return nil, fmt.Errorf("Last-Modified header not provided")
	}
	Expires, ok := headers[common.HeaderExpires]
	if !ok {
		return nil, fmt.Errorf("Expires header not provided")
	}
	return &CompressedResponse{
		RawResp,
		CacheControl,
		ETag,
		LastModified,
		Expires,
	}, nil
}

func (cr *CompressedResponse) BinaryString() (string, error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(cr)
	return b.String(), err
}

func (cr *CompressedResponse) Headers() map[string]string {
	h := make(map[string]string)
	h[common.HeaderETag] = cr.ETag
	h[common.HeaderExpires] = cr.Expires
	h[common.HeaderLastModified] = cr.LastModified
	h[common.HeaderCacheControl] = cr.CacheControl
	return h
}
