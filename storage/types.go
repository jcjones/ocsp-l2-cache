// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package storage

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"golang.org/x/crypto/ocsp"
)

type DocumentType int

type RemoteCache interface {
	Exists(ctx context.Context, key string) (bool, error)
	ExpireAt(ctx context.Context, key string, aExpTime time.Time) error
	SetIfNotExist(ctx context.Context, k string, v string, life time.Duration) (string, error)
	Set(ctx context.Context, k string, v string, life time.Duration) error
	Get(ctx context.Context, k string) (string, bool, error)
	KeysToChan(ctx context.Context, pattern string, c chan<- string) error
}

type Issuer struct {
	spki SPKI
}

func (o Issuer) String() string {
	return o.spki.String()
}

func NewIssuerFromRequest(aReq *ocsp.Request) Issuer {
	obj := Issuer{
		spki: SPKI(aReq.IssuerKeyHash),
	}
	return obj
}

func NewIssuerFromHexKeyId(s string) (*Issuer, error) {
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if len(decoded) != 20 {
		return nil, fmt.Errorf("Key IDs are 20 bytes")
	}
	return &Issuer{
		spki: SPKI(decoded),
	}, nil
}

type SPKI []byte

func (o SPKI) ID() string {
	return base64.URLEncoding.EncodeToString(o)
}

func (o SPKI) String() string {
	return hex.EncodeToString(o)
}

type Serial struct {
	serial []byte
}

type tbsCertWithRawSerial struct {
	Raw          asn1.RawContent
	Version      asn1.RawValue `asn1:"optional,explicit,default:0,tag:0"`
	SerialNumber asn1.RawValue
}

func NewSerial(aCert *x509.Certificate) Serial {
	var tbsCert tbsCertWithRawSerial
	_, err := asn1.Unmarshal(aCert.RawTBSCertificate, &tbsCert)
	if err != nil {
		panic(err)
	}
	return NewSerialFromBytes(tbsCert.SerialNumber.Bytes)
}

func NewSerialFromBytes(b []byte) Serial {
	obj := Serial{
		serial: b,
	}
	return obj
}

func NewSerialFromBigInt(b *big.Int) (Serial, error) {
	if b == nil {
		return Serial{}, fmt.Errorf("null big int")
	}
	obj := NewSerialFromBytes(b.Bytes())
	return obj, nil
}

func NewSerialFromHex(s string) Serial {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return Serial{
		serial: b,
	}
}

func NewSerialFromIDString(s string) (Serial, error) {
	bytes, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return Serial{}, err
	}
	return NewSerialFromBytes(bytes), nil
}

func NewSerialFromBinaryString(s string) (Serial, error) {
	bytes := []byte(s)
	return NewSerialFromBytes(bytes), nil
}

func (s Serial) ID() string {
	return base64.URLEncoding.EncodeToString(s.serial)
}

func (s Serial) String() string {
	return s.HexString()
}

func (s Serial) BinaryString() string {
	return string(s.serial)
}

func (s Serial) HexString() string {
	return hex.EncodeToString(s.serial)
}

func (s Serial) Cmp(o Serial) int {
	return bytes.Compare(s.serial, o.serial)
}

func (s Serial) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.HexString())
}

func (s *Serial) UnmarshalJSON(data []byte) error {
	if data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("Expected surrounding quotes")
	}
	b, err := hex.DecodeString(string(data[1 : len(data)-1]))
	s.serial = b
	return err
}

func (s Serial) MarshalBinary() ([]byte, error) {
	return s.MarshalJSON()
}

func (s *Serial) UnmarshalBinary(data []byte) error {
	return s.UnmarshalJSON(data)
}

func (s *Serial) AsBigInt() *big.Int {
	serialBigInt := big.NewInt(0)
	serialBigInt.SetBytes(s.serial)
	return serialBigInt
}
