package types

import "github.com/tendermint/tendermint/crypto"

type PublicKeyCodec interface {
	Decode(key *PublicKey) (crypto.PubKey, error)
	Encode(key crypto.PubKey) (*PublicKey, error)
}

type publicKeyCodecCacheMiddleware struct {
	cdc PublicKeyCodec
}

func CacheWrapCodec(cdc PublicKeyCodec) PublicKeyCodec {
	return publicKeyCodecCacheMiddleware{cdc: cdc}
}

var _ PublicKeyCodec = publicKeyCodecCacheMiddleware{}

func (p publicKeyCodecCacheMiddleware) Decode(key *PublicKey) (crypto.PubKey, error) {
	res, err := p.cdc.Decode(key)
	if err != nil {
		return nil, err
	}
	key.cachedValue = res
	return res, nil
}

func (p publicKeyCodecCacheMiddleware) Encode(key crypto.PubKey) (*PublicKey, error) {
	res, err := p.cdc.Encode(key)
	if err != nil {
		return nil, err
	}
	res.cachedValue = key
	return res, nil
}
