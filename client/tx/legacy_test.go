package tx_test

import (
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types"
	typestx "github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	memo          = "waboom"
	timeoutHeight = uint64(5)
)

var (
	_, pub1, addr1 = testdata.KeyTestPubAddr()
	_, _, addr2    = testdata.KeyTestPubAddr()
	rawSig         = []byte("dummy")
	msg1           = banktypes.NewMsgSend(addr1, addr2, types.NewCoins(types.NewInt64Coin("wack", 2)))

	chainID = "test-chain"
	tip     = &typestx.Tip{Tipper: addr1.String(), Amount: testdata.NewTestFeeAmount()}
)
