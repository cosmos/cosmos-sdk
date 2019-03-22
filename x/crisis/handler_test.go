package crisis

import (
	"testing"

	skeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

func TestHandleMsgVerifyInvariantWithNotEnoughCoins(t *testing.T) {
	ctx, _, _ := skeeper.CreateTestInput(t, false, 10)
}

func TestHandleMsgVerifyInvariantWithBadInvariant(t *testing.T) {
	// TODO
}

func TestHandleMsgVerifyInvariantWithInvariantNotBroken(t *testing.T) {
	// TODO
}

func TestHandleMsgVerifyInvariantWithInvariantBroken(t *testing.T) {
	// TODO
}
