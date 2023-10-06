// Copyright (c) 2009 The Go Authors. All rights reserved.
// Copyright (c) 2018 Stanislav Fomichev.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//   - Redistributions of source code must retain the above copyright
//
// notice, this list of conditions and the following disclaimer.
//   - Redistributions in binary form must reproduce the above
//
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//   - Neither the name of Google Inc. nor the names of its
//
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
package secp256k1_go

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"testing"
)

var privKey, _ = newPrivFromHex("18E14A7B6A307F426A94F8114701E7C8E774E7F9A47E2C2035DB29A206321725")
var msg, _ = hex.DecodeString("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")

func newPrivFromHex(s string) (ecdsa.PrivateKey, error) {
	k, err := hex.DecodeString(s)
	if err != nil {
		return ecdsa.PrivateKey{}, err
	}

	x, y := SECP256K1().ScalarBaseMult(k)

	return ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: SECP256K1(),
			X:     x,
			Y:     y,
		},
		D: new(big.Int).SetBytes(k),
	}, nil
}

func BenchmarkSign(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := ecdsa.Sign(rand.Reader, &privKey, msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerify(b *testing.B) {
	r, s, err := ecdsa.Sign(rand.Reader, &privKey, msg)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !ecdsa.Verify(&privKey.PublicKey, msg, r, s) {
			b.Fatal("failed to verify signature")
		}
	}
}
