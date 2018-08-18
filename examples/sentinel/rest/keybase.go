package rest

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeyOutput struct {
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	Address sdk.AccAddress `json:"address"`
	PubKey  string         `json:"pub_key"`
	Seed    string         `json:"seed,omitempty"`
}

func getSeed(algo keys.SigningAlgo) string {
	kb := client.MockKeyBase()
	pass := "throwing-this-key-away"
	name := "inmemorykey"
	_, seed, _ := kb.CreateMnemonic(name, keys.English, pass, algo)
	return seed
}
func Bech32KeyOutput(info keys.Info) (KeyOutput, error) {
	account := sdk.AccAddress(info.GetPubKey().Address().Bytes())
	bechPubKey, err := sdk.Bech32ifyAccPub(info.GetPubKey())
	if err != nil {
		return KeyOutput{}, err
	}
	return KeyOutput{
		Name:    info.GetName(),
		Type:    info.GetType(),
		Address: account,
		PubKey:  bechPubKey,
	}, nil
}
