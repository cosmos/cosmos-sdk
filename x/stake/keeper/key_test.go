package keeper

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var (
	pk1 = ed25519.GenPrivKey().PubKey()
)

func TestGetValidatorPowerRank(t *testing.T) {
	valAddr1 := sdk.ValAddress(pk1.Bytes())
	emptyDesc := types.Description{}
	val1 := types.NewValidator(valAddr1, pk1, emptyDesc)
	val1.Tokens = sdk.NewDec(12)

	tests := []struct {
		validator types.Validator
		wantHex   string
	}{
		{val1, "05303030303030303030303132ffffffffffffffffffff"},
	}
	for i, tt := range tests {
		got := hex.EncodeToString(getValidatorPowerRank(tt.validator))

		assert.Equal(t, tt.wantHex, got, "Keys did not match on test case %d", i)
	}
}
