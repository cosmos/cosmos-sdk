package params_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func testProposal(changes ...params.Change) params.ProposalChange {
	return params.NewProposalChange(
		"Test",
		"description",
		"myspace",
		changes,
	)
}
func TestProposalPassedEndblocker(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := gov.GetMockApp(t, 1)

	tp := testProposal(params.NewChange([]byte{0x00}, nil, "mychange"))
	resTags := gov.TestProposal(t, mapp, addrs[0], keeper, sdk, testProposal)
}
