// +build norace

package client_test

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (s *IntegrationTestSuite) TestQueryABCIHeight() {
	// test ABCI query uses request height argument
	// instead of client context height
	contextHeight := 1 // query at height 1 or 2 would cause an error
	reqHeight := int64(10)

	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	clientCtx = clientCtx.WithHeight(contextHeight)

	req := abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", banktypes.StoreKey),
		Height: reqHeight,
		Data:   banktypes.CreateAccountBalancesPrefix(val.Address),
		Prove:  true,
	}

	res, err := clientCtx.QueryABCI(req)
	s.Require().NoError(err)

	s.Require().Equal(reqHeight, res.Height)

}
