package tx_test

import (
	banktypes "cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types"
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
)
