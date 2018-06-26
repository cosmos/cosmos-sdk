package hkdfchacha20poly1305

import (
	"bytes"
	cr "crypto/rand"
	"encoding/hex"
	mr "math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test that a test vector we generated is valid. (Ensures backwards
// compatability)
func TestVector(t *testing.T) {
	key, _ := hex.DecodeString("56f8de45d3c294c7675bcaf457bdd4b71c380b9b2408ce9412b348d0f08b69ee")
	aead, err := New(key[:])
	if err != nil {
		t.Fatal(err)
	}
	cts := []string{"e20a8bf42c535ac30125cfc52031577f0b",
		"657695b37ba30f67b25860d90a6f1d00d8",
		"e9aa6f3b7f625d957fd50f05bcdf20d014",
		"8a00b3b5a6014e0d2033bebc5935086245",
		"aadd74867b923879e6866ea9e03c009039",
		"fc59773c2c864ee3b4cc971876b3c7bed4",
		"caec14e3a9a52ce1a2682c6737defa4752",
		"0b89511ffe490d2049d6950494ee51f919",
		"7de854ea71f43ca35167a07566c769083d",
		"cd477327f4ea4765c71e311c5fec1edbfb"}

	for i := 0; i < 10; i++ {
		ct, _ := hex.DecodeString(cts[i])

		byteArr := []byte{byte(i)}
		nonce := make([]byte, 24, 24)
		nonce[0] = byteArr[0]

		plaintext, err := aead.Open(nil, nonce, ct, byteArr)
		if err != nil {
			t.Errorf("%dth Open failed", i)
			continue
		}
		assert.Equal(t, byteArr, plaintext)
	}
}

// The following test is taken from
// https://github.com/golang/crypto/blob/master/chacha20poly1305/chacha20poly1305_test.go#L69
// It requires the below copyright notice, where "this source code" refers to the following function.
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found at the bottom of this file.
func TestRandom(t *testing.T) {
	// Some random tests to verify Open(Seal) == Plaintext
	for i := 0; i < 256; i++ {
		var nonce [24]byte
		var key [32]byte

		al := mr.Intn(128)
		pl := mr.Intn(16384)
		ad := make([]byte, al)
		plaintext := make([]byte, pl)
		cr.Read(key[:])
		cr.Read(nonce[:])
		cr.Read(ad)
		cr.Read(plaintext)

		aead, err := New(key[:])
		if err != nil {
			t.Fatal(err)
		}

		ct := aead.Seal(nil, nonce[:], plaintext, ad)

		plaintext2, err := aead.Open(nil, nonce[:], ct, ad)
		if err != nil {
			t.Errorf("Random #%d: Open failed", i)
			continue
		}

		if !bytes.Equal(plaintext, plaintext2) {
			t.Errorf("Random #%d: plaintext's don't match: got %x vs %x", i, plaintext2, plaintext)
			continue
		}

		if len(ad) > 0 {
			alterAdIdx := mr.Intn(len(ad))
			ad[alterAdIdx] ^= 0x80
			if _, err := aead.Open(nil, nonce[:], ct, ad); err == nil {
				t.Errorf("Random #%d: Open was successful after altering additional data", i)
			}
			ad[alterAdIdx] ^= 0x80
		}

		alterNonceIdx := mr.Intn(aead.NonceSize())
		nonce[alterNonceIdx] ^= 0x80
		if _, err := aead.Open(nil, nonce[:], ct, ad); err == nil {
			t.Errorf("Random #%d: Open was successful after altering nonce", i)
		}
		nonce[alterNonceIdx] ^= 0x80

		alterCtIdx := mr.Intn(len(ct))
		ct[alterCtIdx] ^= 0x80
		if _, err := aead.Open(nil, nonce[:], ct, ad); err == nil {
			t.Errorf("Random #%d: Open was successful after altering ciphertext", i)
		}
		ct[alterCtIdx] ^= 0x80
	}
}

// AFOREMENTIONED LICENCE
// Copyright (c) 2009 The Go Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
