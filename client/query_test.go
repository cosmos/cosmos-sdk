// +build norace

package client_test

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (s *IntegrationTestSuite) TestQueryABCIHeight() {
	testCases := []struct {
		name      string
		reqHeight int64
		ctxHeight int64
		expHeight int64
	}{
		{
			name:      "non zero request height",
			reqHeight: 3,
			ctxHeight: 1, // query at height 1 or 2 would cause an error
			expHeight: 3,
		},
		{
			name:      "empty request height - use context height",
			reqHeight: 0,
			ctxHeight: 3,
			expHeight: 3,
		},
		{
			name:      "empty request height and context height - use latest height",
			reqHeight: 0,
			ctxHeight: 0,
			expHeight: 4,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.network.WaitForHeight(tc.expHeight)

			val := s.network.Validators[0]

			clientCtx := val.ClientCtx
			clientCtx = clientCtx.WithHeight(tc.ctxHeight)

			req := abci.RequestQuery{
				Path:   fmt.Sprintf("store/%s/key", banktypes.StoreKey),
				Height: tc.reqHeight,
				Data:   append(banktypes.BalancesPrefix, val.Address.Bytes()...),
				Prove:  true,
			}

			res, err := clientCtx.QueryABCI(req)
			s.Require().NoError(err)

			s.Require().Equal(tc.expHeight, res.Height)
		})
	}
}
