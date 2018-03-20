package stake

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	wire "github.com/tendermint/go-wire"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	validator = []byte("addressvalidator1")
	empty     sdk.Address

	coinPos          = sdk.Coin{"fermion", 1000}
	coinZero         = sdk.Coin{"fermion", 0}
	coinNeg          = sdk.Coin{"fermion", -10000}
	coinPosNotAtoms  = sdk.Coin{"foo", 10000}
	coinZeroNotAtoms = sdk.Coin{"foo", 0}
	coinNegNotAtoms  = sdk.Coin{"foo", -10000}
)

//TODO add these tests to one of some of the types
//func TestMsgAddrValidateBasic(t *testing.T) {
//tests := []struct {
//name    string
//address sdk.Address
//wantErr bool
//}{
//{"basic good", addrs[0], false},
//{"empty delegator", sdk.Address{}, true},
//}

//for _, tc := range tests {
//tx := NewMsgAddr(tc.address)
//assert.Equal(t, tc.wantErr, tx.ValidateBasic() != nil,
//"test: %v, tx.ValidateBasic: %v", tc.name, tx.ValidateBasic())
//}
//}

//func TestValidateCoin(t *testing.T) {
//tests := []struct {
//name    string
//coin    sdk.Coin
//wantErr bool
//}{
//{"basic good", coinPos, false},
//{"zero coin", coinZero, true},
//{"neg coin", coinNeg, true},
//}

//for _, tc := range tests {
//assert.Equal(t, tc.wantErr, validateCoin(tc.coin) != nil,
//"test: %v, tx.ValidateBasic: %v", tc.name, validateCoin(tc.coin))
//}
//}

func TestSerializeMsg(t *testing.T) {

	// make sure all types construct properly
	bondAmt := 1234321
	bond := sdk.Coin{Denom: "atom", Amount: int64(bondAmt)}

	tests := []struct {
		tx sdk.Msg
	}{
		{NewMsgDeclareCandidacy(addrs[0], pks[0], bond, Description{})},
		{NewMsgEditCandidacy(addrs[0], Description{})},
		{NewMsgDelegate(addrs[0], addrs[1], bond)},
		{NewMsgUnbond(addrs[0], addrs[1], strconv.Itoa(bondAmt))},
	}

	for i, tc := range tests {
		var tx sdk.Tx
		bs := wire.BinaryBytes(tc.tx)
		err := wire.ReadBinaryBytes(bs, &tx)
		if assert.NoError(t, err, "%d", i) {
			assert.Equal(t, tc.tx, tx, "%d", i)
		}
	}
}
