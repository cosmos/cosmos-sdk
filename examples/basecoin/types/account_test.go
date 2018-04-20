package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

func TestAppAccount(t *testing.T) {
	cdc := wire.NewCodec()
	addr := sdk.Address([]byte("address"))
	acc := &AppAccount{
		BaseAccount: auth.BaseAccount{
			Address: addr,
			Coins:   sdk.Coins{},
		},
		Name: "",
	}

	bz, err := cdc.MarshalBinary(acc)
	assert.Nil(t, err)

	decode := GetAccountDecoder(cdc)
	res, err := decode(bz)
	assert.Nil(t, err)
	assert.Equal(t, acc, res)

	name := t.Name()
	acc.SetName(name)
	accname := acc.GetName()
	assert.Equal(t, name, accname)
}
