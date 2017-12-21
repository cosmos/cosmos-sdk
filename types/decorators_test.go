package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/stretchr/testify/assert"
)

func TestDecorate(t *testing.T) {

	var calledDec1, calledDec2, calledHandler bool
	dec1 := func(ctx Context, ms store.MultiStore, tx Tx, next Handler) Result {
		calledDec1 = true
		next(ctx, ms, tx)
		return Result{}
	}

	dec2 := func(ctx Context, ms store.MultiStore, tx Tx, next Handler) Result {
		calledDec2 = true
		next(ctx, ms, tx)
		return Result{}
	}

	handler := func(ctx Context, ms store.MultiStore, tx Tx) Result {
		calledHandler = true
		return Result{}
	}

	decoratedHandler := ChainDecorators(dec1, dec2).WithHandler(handler)

	var ctx Context
	var ms store.MultiStore
	var tx Tx
	decoratedHandler(ctx, ms, tx)
	assert.True(t, calledDec1)
	assert.True(t, calledDec2)
	assert.True(t, calledHandler)
}
