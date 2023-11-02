package tx_test

import (
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	countertypes "github.com/cosmos/cosmos-sdk/x/counter/types"
)

const (
	memo          = "waboom"
	timeoutHeight = uint64(5)
)

var (
	_, pub1, addr1 = testdata.KeyTestPubAddr()
	_, _, addr2    = testdata.KeyTestPubAddr()
	rawSig         = []byte("dummy")
	msg1           = &countertypes.MsgIncreaseCounter{Signer: addr1.String(), Count: 1}

	chainID = "test-chain"
)
