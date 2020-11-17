package keyring

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/internal/protocdc"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: Move this file to client/keys

// KeyOutput defines a structure wrapping around an Info object used for output
// functionality.
type KeyOutput struct {
	Name      string                 `json:"name" yaml:"name"`
	Type      string                 `json:"type" yaml:"type"`
	Address   string                 `json:"address" yaml:"address"`
	PubKey    string                 `json:"pubkey" yaml:"pubkey"`
	Mnemonic  string                 `json:"mnemonic,omitempty" yaml:"mnemonic"`
	Threshold uint                   `json:"threshold,omitempty" yaml:"threshold"`
	PubKeys   []multisigPubKeyOutput `json:"pubkeys,omitempty" yaml:"pubkeys"`
}

// NewKeyOutput creates a default KeyOutput instance without Mnemonic, Threshold and PubKeys
func NewKeyOutput(name string, keyType KeyType, a sdk.Address, pk cryptotypes.PubKey) (KeyOutput, error) {
	pkBytes, err := protocdc.MarshalJSON(pk, nil)
	if err != nil {
		return KeyOutput{}, err
	}
	return KeyOutput{
		Name:    name,
		Type:    keyType.String(),
		Address: a.String(),
		PubKey:  string(pkBytes),
	}, nil
}

type multisigPubKeyOutput struct {
	Address string `json:"address" yaml:"address"`
	PubKey  string `json:"pubkey" yaml:"pubkey"`
	Weight  uint   `json:"weight" yaml:"weight"`
}

// Bech32KeysOutput returns a slice of KeyOutput objects, each with the "acc"
// Bech32 prefixes, given a slice of Info objects. It returns an error if any
// call to Bech32KeyOutput fails.
func Bech32KeysOutput(infos []Info) ([]KeyOutput, error) {
	kos := make([]KeyOutput, len(infos))
	var err error
	for i, info := range infos {
		kos[i], err = Bech32KeyOutput(info)
		if err != nil {
			return nil, err
		}
	}

	return kos, nil
}

// Bech32ConsKeyOutput create a KeyOutput in with "cons" Bech32 prefixes.
func Bech32ConsKeyOutput(keyInfo Info) (KeyOutput, error) {
	pk := keyInfo.GetPubKey()
	addr := sdk.ConsAddress(pk.Address().Bytes())
	return NewKeyOutput(keyInfo.GetName(), keyInfo.GetType(), addr, pk)
}

// Bech32ValKeyOutput create a KeyOutput in with "val" Bech32 prefixes.
func Bech32ValKeyOutput(keyInfo Info) (KeyOutput, error) {
	pk := keyInfo.GetPubKey()
	addr := sdk.ValAddress(pk.Address().Bytes())
	return NewKeyOutput(keyInfo.GetName(), keyInfo.GetType(), addr, pk)
}

// Bech32KeyOutput create a KeyOutput in with "acc" Bech32 prefixes. If the
// public key is a multisig public key, then the threshold and constituent
// public keys will be added.
func Bech32KeyOutput(keyInfo Info) (KeyOutput, error) {
	pk := keyInfo.GetPubKey()
	addr := sdk.AccAddress(pk.Address().Bytes())
	ko, err := NewKeyOutput(keyInfo.GetName(), keyInfo.GetType(), addr, pk)
	if err != nil {
		return ko, err
	}

	if mInfo, ok := keyInfo.(*multiInfo); ok {
		pubKeys := make([]multisigPubKeyOutput, len(mInfo.PubKeys))

		for i, pkInfo := range mInfo.PubKeys {
			pk = pkInfo.PubKey
			addr = sdk.AccAddress(pk.Address().Bytes())
			pkBytes, err := protocdc.MarshalJSON(pk, nil)
			if err != nil {
				return ko, err
			}
			pubKeys[i] = multisigPubKeyOutput{addr.String(), string(pkBytes), pkInfo.Weight}
		}

		ko.Threshold = mInfo.Threshold
		ko.PubKeys = pubKeys
	}

	return ko, nil
}
