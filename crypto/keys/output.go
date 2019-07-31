package keys

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
func NewKeyOutput(name, keyType, address, pubkey string) KeyOutput {
	return KeyOutput{
		Name:    name,
		Type:    keyType,
		Address: address,
		PubKey:  pubkey,
	}
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
	for i, info := range infos {
		ko, err := Bech32KeyOutput(info)
		if err != nil {
			return nil, err
		}
		kos[i] = ko
	}

	return kos, nil
}

// Bech32ConsKeyOutput create a KeyOutput in with "cons" Bech32 prefixes.
func Bech32ConsKeyOutput(keyInfo Info) (KeyOutput, error) {
	consAddr := sdk.ConsAddress(keyInfo.GetPubKey().Address().Bytes())

	bechPubKey, err := sdk.Bech32ifyConsPub(keyInfo.GetPubKey())
	if err != nil {
		return KeyOutput{}, err
	}

	return NewKeyOutput(keyInfo.GetName(), keyInfo.GetType().String(), consAddr.String(), bechPubKey), nil
}

// Bech32ValKeyOutput create a KeyOutput in with "val" Bech32 prefixes.
func Bech32ValKeyOutput(keyInfo Info) (KeyOutput, error) {
	valAddr := sdk.ValAddress(keyInfo.GetPubKey().Address().Bytes())

	bechPubKey, err := sdk.Bech32ifyValPub(keyInfo.GetPubKey())
	if err != nil {
		return KeyOutput{}, err
	}

	return NewKeyOutput(keyInfo.GetName(), keyInfo.GetType().String(), valAddr.String(), bechPubKey), nil
}

// Bech32KeyOutput create a KeyOutput in with "acc" Bech32 prefixes. If the
// public key is a multisig public key, then the threshold and constituent
// public keys will be added.
func Bech32KeyOutput(keyInfo Info) (KeyOutput, error) {
	accAddr := sdk.AccAddress(keyInfo.GetPubKey().Address().Bytes())
	bechPubKey, err := sdk.Bech32ifyAccPub(keyInfo.GetPubKey())
	if err != nil {
		return KeyOutput{}, err
	}

	ko := NewKeyOutput(keyInfo.GetName(), keyInfo.GetType().String(), accAddr.String(), bechPubKey)

	if mInfo, ok := keyInfo.(*multiInfo); ok {
		pubKeys := make([]multisigPubKeyOutput, len(mInfo.PubKeys))

		for i, pk := range mInfo.PubKeys {
			accAddr := sdk.AccAddress(pk.PubKey.Address().Bytes())

			bechPubKey, err := sdk.Bech32ifyAccPub(pk.PubKey)
			if err != nil {
				return KeyOutput{}, err
			}

			pubKeys[i] = multisigPubKeyOutput{accAddr.String(), bechPubKey, pk.Weight}
		}

		ko.Threshold = mInfo.Threshold
		ko.PubKeys = pubKeys
	}

	return ko, nil
}
