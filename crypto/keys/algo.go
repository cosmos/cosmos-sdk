package keys

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	tmed25519 "github.com/tendermint/tendermint/crypto/ed25519"
	tmsm2 "github.com/tendermint/tendermint/crypto/sm2"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/sm2"
)

func FromTmPubKey(pubkey crypto.PubKey) (crypto.PubKey, error) {
	switch pubkey.(type) {
	case tmed25519.PubKey:
		return &ed25519.PubKey{Key: pubkey.Bytes()}, nil
	case tmsm2.PubKeySm2:
		return &sm2.PubKey{Key: pubkey.Bytes()}, nil
	}
	return nil, fmt.Errorf("not support pubkey type: %s", pubkey.Type())
}
