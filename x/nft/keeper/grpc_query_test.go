package keeper_test

import (
	gocontext "context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/x/nft"
)

func TestGRPCQuery(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) TestGRPCQueryBalance() {
	var (
		req     *nft.QueryBalanceRequest
		balance uint64
	)
	testCases := []struct {
		msg      string
		malleate func(require *require.Assertions)
		expError string
		postTest func(require *require.Assertions, res *nft.QueryBalanceResponse)
	}{
		{
			"fail empty ClassId",
			func(require *require.Assertions) {
				req = &nft.QueryBalanceRequest{}
			},
			"class id can not be empty",
			func(require *require.Assertions, res *nft.QueryBalanceResponse) {},
		},
		{
			"fail invalid Owner addr",
			func(require *require.Assertions) {
				req = &nft.QueryBalanceRequest{
					ClassId: testClassID,
					Owner:   "owner",
				}
			},
			"decoding bech32 failed",
			func(require *require.Assertions, res *nft.QueryBalanceResponse) {},
		},
		{
			"Success",
			func(require *require.Assertions) {
				suite.TestMint()
				req = &nft.QueryBalanceRequest{
					ClassId: testClassID,
					Owner:   suite.addrs[0].String(),
				}
				balance = 1
			},
			"",
			func(require *require.Assertions, res *nft.QueryBalanceResponse) {
				require.Equal(res.Amount, balance)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := suite.Require()
			tc.malleate(require)
			result, err := suite.queryClient.Balance(gocontext.Background(), req)
			if tc.expError == "" {
				require.NoError(err)
			} else {
				require.Error(err)
				require.Contains(err.Error(), tc.expError)
			}
			tc.postTest(require, result)
		})
	}
}

func (suite *TestSuite) TestGRPCQueryOwner() {
	var (
		req   *nft.QueryOwnerRequest
		owner string
	)
	testCases := []struct {
		msg      string
		malleate func(require *require.Assertions)
		expError string
		postTest func(require *require.Assertions, res *nft.QueryOwnerResponse)
	}{
		{
			"fail empty ClassId",
			func(require *require.Assertions) {
				req = &nft.QueryOwnerRequest{
					Id: testID,
				}
			},
			"class id can not be empty",
			func(require *require.Assertions, res *nft.QueryOwnerResponse) {},
		},
		{
			"fail empty nft id",
			func(require *require.Assertions) {
				req = &nft.QueryOwnerRequest{
					ClassId: testClassID,
				}
			},
			"nft id can not be empty",
			func(require *require.Assertions, res *nft.QueryOwnerResponse) {},
		},
		{
			"success but nft id not exist",
			func(require *require.Assertions) {
				req = &nft.QueryOwnerRequest{
					ClassId: testClassID,
					Id:      "kitty2",
				}
			},
			"",
			func(require *require.Assertions, res *nft.QueryOwnerResponse) {
				require.Equal(res.Owner, owner)
			},
		},
		{
			"success but class id not exist",
			func(require *require.Assertions) {
				req = &nft.QueryOwnerRequest{
					ClassId: "kitty1",
					Id:      testID,
				}
			},
			"",
			func(require *require.Assertions, res *nft.QueryOwnerResponse) {
				require.Equal(res.Owner, owner)
			},
		},
		{
			"Success",
			func(require *require.Assertions) {
				suite.TestMint()
				req = &nft.QueryOwnerRequest{
					ClassId: testClassID,
					Id:      testID,
				}
				owner = suite.addrs[0].String()
			},
			"",
			func(require *require.Assertions, res *nft.QueryOwnerResponse) {
				require.Equal(res.Owner, owner)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := suite.Require()
			tc.malleate(require)
			result, err := suite.queryClient.Owner(gocontext.Background(), req)
			if tc.expError == "" {
				require.NoError(err)
			} else {
				require.Error(err)
				require.Contains(err.Error(), tc.expError)
			}
			tc.postTest(require, result)
		})
	}
}

func (suite *TestSuite) TestGRPCQuerySupply() {
	var (
		req    *nft.QuerySupplyRequest
		supply uint64
	)
	testCases := []struct {
		msg      string
		malleate func(require *require.Assertions)
		expError string
		postTest func(require *require.Assertions, res *nft.QuerySupplyResponse)
	}{
		{
			"fail empty ClassId",
			func(require *require.Assertions) {
				req = &nft.QuerySupplyRequest{}
			},
			"class id can not be empty",
			func(require *require.Assertions, res *nft.QuerySupplyResponse) {},
		},
		{
			"success but class id not exist",
			func(require *require.Assertions) {
				req = &nft.QuerySupplyRequest{
					ClassId: "kitty1",
				}
			},
			"",
			func(require *require.Assertions, res *nft.QuerySupplyResponse) {
				require.Equal(res.Amount, supply)
			},
		},
		{
			"success but supply equal zero",
			func(require *require.Assertions) {
				req = &nft.QuerySupplyRequest{
					ClassId: testClassID,
				}
				suite.TestSaveClass()
			},
			"",
			func(require *require.Assertions, res *nft.QuerySupplyResponse) {
				require.Equal(res.Amount, supply)
			},
		},
		{
			"Success",
			func(require *require.Assertions) {
				n := nft.NFT{
					ClassId: testClassID,
					Id:      testID,
					Uri:     testURI,
				}
				err := suite.app.NFTKeeper.Mint(suite.ctx, n, suite.addrs[0])
				suite.Require().NoError(err)

				req = &nft.QuerySupplyRequest{
					ClassId: testClassID,
				}
				supply = 1
			},
			"",
			func(require *require.Assertions, res *nft.QuerySupplyResponse) {
				require.Equal(res.Amount, supply)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := suite.Require()
			tc.malleate(require)
			result, err := suite.queryClient.Supply(gocontext.Background(), req)
			if tc.expError == "" {
				require.NoError(err)
			} else {
				require.Error(err)
				require.Contains(err.Error(), tc.expError)
			}
			tc.postTest(require, result)
		})
	}
}

func (suite *TestSuite) TestGRPCQueryNFTsOfClass() {
	var (
		req  *nft.QueryNFTsOfClassRequest
		nfts []*nft.NFT
	)
	testCases := []struct {
		msg      string
		malleate func(require *require.Assertions)
		expError string
		postTest func(require *require.Assertions, res *nft.QueryNFTsOfClassResponse)
	}{
		{
			"fail empty ClassId",
			func(require *require.Assertions) {
				req = &nft.QueryNFTsOfClassRequest{}
			},
			"class id can not be empty",
			func(require *require.Assertions, res *nft.QueryNFTsOfClassResponse) {},
		},
		{
			"success but class id not exist",
			func(require *require.Assertions) {
				req = &nft.QueryNFTsOfClassRequest{
					ClassId: "kitty1",
				}
			},
			"",
			func(require *require.Assertions, res *nft.QueryNFTsOfClassResponse) {
				require.Equal(res.Amount, supply)
			},
		},
		{
			"success but supply equal zero",
			func(require *require.Assertions) {
				req = &nft.QueryNFTsOfClassRequest{
					ClassId: testClassID,
				}
				suite.TestSaveClass()
			},
			"",
			func(require *require.Assertions, res *nft.QueryNFTsOfClassResponse) {
				require.Equal(res.Amount, supply)
			},
		},
		{
			"Success",
			func(require *require.Assertions) {
				n := nft.NFT{
					ClassId: testClassID,
					Id:      testID,
					Uri:     testURI,
				}
				err := suite.app.NFTKeeper.Mint(suite.ctx, n, suite.addrs[0])
				suite.Require().NoError(err)

				req = &nft.QueryNFTsOfClassRequest{
					ClassId: testClassID,
				}
				supply = 1
			},
			"",
			func(require *require.Assertions, res *nft.QueryNFTsOfClassResponse) {
				require.Equal(res.Amount, supply)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := suite.Require()
			tc.malleate(require)
			result, err := suite.queryClient.NFTsOfClass(gocontext.Background(), req)
			if tc.expError == "" {
				require.NoError(err)
			} else {
				require.Error(err)
				require.Contains(err.Error(), tc.expError)
			}
			tc.postTest(require, result)
		})
	}
}
