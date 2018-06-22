package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestContextWithSigners(t *testing.T) {
	ms, _, _ := setupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "mychainid"}, false, log.NewNopLogger())

	_, _, addr1 := keyPubAddr()
	_, _, addr2 := keyPubAddr()
	acc1 := NewBaseAccountWithAddress(addr1)
	acc1.SetSequence(7132)
	acc2 := NewBaseAccountWithAddress(addr2)
	acc2.SetSequence(8821)

	// new ctx has no signers
	signers := GetSigners(ctx)
	assert.Equal(t, 0, len(signers))

	ctx2 := WithSigners(ctx, []Account{&acc1, &acc2})

	// original context is unchanged
	signers = GetSigners(ctx)
	assert.Equal(t, 0, len(signers))

	// new context has signers
	signers = GetSigners(ctx2)
	assert.Equal(t, 2, len(signers))
	assert.Equal(t, acc1, *(signers[0].(*BaseAccount)))
	assert.Equal(t, acc2, *(signers[1].(*BaseAccount)))
}
