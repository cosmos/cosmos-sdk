package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	// wire "github.com/tendermint/go-amino"
)

func buildTx(msg sdk.Msg, name string) ([]byte, error) {
	keybase, err := keys.GetKeyBase()
	if err != nil {
		return nil, err
	}

	bz := msg.GetSignBytes()
	buf := client.BufferStdin()
	prompt := fmt.Sprintf("Password to sign with '%s':", name)
	passphrase, err := client.GetPassword(prompt, buf)
	if err != nil {
		return nil, err
	}
	sig, pubkey, err := keybase.Sign(name, passphrase, bz)
	if err != nil {
		return nil, err
	}
	sigs := []sdk.StdSignature{{
		PubKey:    pubkey,
		Signature: sig,
		Sequence:  viper.GetInt64(flagSequence),
	}}

	tx := sdk.NewStdTx(msg, sigs)

	txBytes, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}
	return txBytes, nil
}

func getAddress(name string) []byte {
	keybase, err := keys.GetKeyBase()
	if err != nil {
		panic(err)
	}

	info, err := keybase.Get(name)
	if err != nil {
		panic(err)
	}

	return info.Address()
}
