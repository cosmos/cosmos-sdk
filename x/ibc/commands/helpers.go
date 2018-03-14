package commands

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

func buildTx(cdc *wire.Codec, msg sdk.Msg) ([]byte, error) {
	keybase, err := keys.GetKeyBase()
	if err != nil {
		return nil, err
	}

	name := viper.GetString(client.FlagName)

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

	txBytes, err := cdc.MarshalBinary(tx)
	if err != nil {
		return nil, err
	}
	return txBytes, nil
}

func getAddress() []byte {
	keybase, err := keys.GetKeyBase()
	if err != nil {
		panic(err)
	}

	name := viper.GetString(client.FlagName)
	info, err := keybase.Get(name)
	if err != nil {
		panic(err)
	}

	return info.Address()
}
