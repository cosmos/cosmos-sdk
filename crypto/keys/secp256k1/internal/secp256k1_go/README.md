# secp256k1 implementation in Go

This is basically Go's src/crypto/elliptic/elliptic.go but with
special additions to make it work with a=0.

For the example usage see `signature_test.go`.

Curve parameters taken from: http://www.secg.org/sec2-v2.pdf

Actual implementation based on:

* http://hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-0.html#addition-add-2007-bl
* http://hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-0.html#doubling-dbl-2009-l
