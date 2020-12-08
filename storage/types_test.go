// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package storage

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"reflect"
	"testing"
)

// issuer:ca
// subject: leadingZeros
// serialNumber: 0x00AA
//
// ... requires hacking pycert.py

const (
	kLeadingZeroes = `-----BEGIN CERTIFICATE-----
MIICozCCAYugAwIBAgICAKowDQYJKoZIhvcNAQELBQAwDTELMAkGA1UEAwwCY2Ew
IhgPMjAxNzExMjcwMDAwMDBaGA8yMDIwMDIwNTAwMDAwMFowGDEWMBQGA1UEAwwN
IGxlYWRpbmdaZXJvczCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALqI
UahEjhbWQf1utogGNhA9PBPZ6uQ1SrTs9WhXbCR7wcclqODYH72xnAabbhqG8mvi
r1p1a2pkcQh6pVqnRYf3HNUknAJ+zUP8HmnQOCApk6sgw0nk27lMwmtsDu0Vgg/x
fq1pGrHTAjqLKkHup3DgDw2N/WYLK7AkkqR9uYhheZCxV5A90jvF4LhIH6g304hD
7ycW2FW3ZlqqfgKQLzp7EIAGJMwcbJetlmFbt+KWEsB1MaMMkd20yvf8rR0l0wnv
uRcOp2jhs3svIm9p47SKlWEd7ibWJZ2rkQhONsscJAQsvxaLL+Xxj5kXMbiz/kkj
+nJRxDHVA6zaGAo17Y0CAwEAATANBgkqhkiG9w0BAQsFAAOCAQEAGGxF47xA91w0
JvJ9kMGyiTqwtU7RaCXW+euVrFq8fFqE6+Gy+EnAQkNvzAjgHBoboodsost7xwuq
JG/LoF6qUsztYVpGHtpElghTv6XXhMCh0zaoM0PrE5oXYY75di+ltEH1DJVf0xj0
30AK23vyZ+UsNwISUyzECxA10RUSAD697vFIqW9RrJG1fM6f3l/VRBLINqOafrNB
z6brFHZzowdAKMBkog7ZQyiHEi1BqV8Vd8SKng2lQNw67RFgfB2Ltgbew2SiZMor
ylxqvBshawlL7jExLaSnMgE0RvcvSjpDguO7QO84CtH2LDGYjBABfy9ShGWTsKHi
Tqhe91GhlQ==
-----END CERTIFICATE-----`
)

func TestIssuerHex(t *testing.T) {
	t.Parallel()
	_, err := NewIssuerFromHexKeyId("what?")
	if err == nil {
		t.Error("not hex, should have failed")
	}

	_, err = NewIssuerFromHexKeyId("abcd")
	if err == nil {
		t.Error("not long enough, should have failed")
	}

	i, err := NewIssuerFromHexKeyId("142EB317B75856CBAE500940E61FAF9D8B14C2C6")
	if err != nil {
		t.Error(err)
	}
	if i.spki.String() != "142eb317b75856cbae500940e61faf9d8b14c2c6" {
		t.Errorf("Unexpected value: %+v", i.spki.String())
	}
}

func TestSerial(t *testing.T) {
	t.Parallel()
	x := NewSerialFromHex("DEADBEEF")
	y := Serial{
		serial: []byte{0xDE, 0xAD, 0xBE, 0xEF},
	}
	if !reflect.DeepEqual(x, y) {
		t.Errorf("Serials should match")
	}

	if x.Cmp(y) != 0 {
		t.Errorf("Should compare the same")
	}

	if y.String() != "deadbeef" {
		t.Errorf("Wrong encoding, got: %s but expected deadbeef", y.String())
	}

	if x.String() != "deadbeef" {
		t.Errorf("Wrong encoding, got: %s but expected deadbeef", y.String())
	}
}

func TestSerialFromCertWithLeadingZeroes(t *testing.T) {
	t.Parallel()
	b, _ := pem.Decode([]byte(kLeadingZeroes))

	cert, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		t.Error(err)
	}

	x := NewSerial(cert)
	// The Serial should be only the Value of the serialNumber field, so in this
	// case [00, AA].
	// The Stringification is the hexification, lowercase
	if x.String() != "00aa" {
		t.Errorf("Lost leading zeroes: %s != 00aa", x.String())
	}

	// The internal ID repr is base64
	if x.ID() != "AKo=" {
		t.Errorf("ID was %s but should be AKo=", x.ID())
	}
}

func TestSerialJson(t *testing.T) {
	t.Parallel()
	serials := []Serial{NewSerialFromHex("ABCDEF"), NewSerialFromHex("001100")}
	data, err := json.Marshal(serials)
	if err != nil {
		t.Error(err)
	}

	var decoded []Serial
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("Decoding %s got error %v", string(data), err)
	}

	if !reflect.DeepEqual(serials, decoded) {
		t.Errorf("Should match %+v %+v", serials, decoded)
	}
}

func TestSerialBigInt(t *testing.T) {
	t.Parallel()
	bint := big.NewInt(0xCAFEDEAD)
	serial := NewSerialFromBytes(bint.Bytes())
	reflex := serial.AsBigInt()
	if reflex.Cmp(bint) != 0 {
		t.Errorf("Expected %v but got %v", bint, reflex)
	}

	bigserial, err := NewSerialFromBigInt(bint)
	if err != nil {
		t.Error(err)
	}
	if bigserial.Cmp(serial) != 0 {
		t.Errorf("Expected %v but got %v", serial, bigserial)
	}
}

func TestSerialBinaryStrings(t *testing.T) {
	t.Parallel()
	serials := []Serial{
		NewSerialFromHex("ABCDEF"),
		NewSerialFromHex("001100"),
		NewSerialFromHex("ABCDEF0100101010010101010100101010"),
		NewSerialFromHex("00ABCDEF01001010101010101010010101"),
		NewSerialFromHex("FFFFFFFFFFFFFF00F00FFFFFFFFFFFFFFF"),
	}

	for _, s := range serials {
		astr := s.BinaryString()

		decoded, err := NewSerialFromBinaryString(astr)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(s, decoded) {
			t.Errorf("Expected to match %v != %v", s, decoded)
		}
	}
}

func TestSerialID(t *testing.T) {
	t.Parallel()
	x := NewSerialFromHex("DEADBEEF")
	idStr := x.ID()
	decoded, err := NewSerialFromIDString(idStr)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(x, decoded) {
		t.Errorf("Should match %+v & %+v", x, decoded)
	}

	if _, err := NewSerialFromIDString("not base64"); err == nil {
		t.Error("Expected an error decoding an invalid ID string")
	}

	if x.HexString() != "deadbeef" {
		t.Errorf("Expected HexString to match %s", x.HexString())
	}
}
