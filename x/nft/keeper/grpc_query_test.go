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
			"success, no nft",
			func(require *require.Assertions) {
				req = &nft.QueryNFTsOfClassRequest{
					ClassId: testClassID,
				}
				suite.TestSaveClass()
			},
			"",
			func(require *require.Assertions, res *nft.QueryNFTsOfClassResponse) {
				require.Len(res.Nfts, 0)
			},
		},
		{
			"success, class id not exist",
			func(require *require.Assertions) {
				req = &nft.QueryNFTsOfClassRequest{
					ClassId: "kitty1",
				}
				n := nft.NFT{
					ClassId: testClassID,
					Id:      testID,
					Uri:     testURI,
				}
				err := suite.app.NFTKeeper.Mint(suite.ctx, n, suite.addrs[0])
				suite.Require().NoError(err)
			},
			"",
			func(require *require.Assertions, res *nft.QueryNFTsOfClassResponse) {
				require.Len(res.Nfts, 0)
			},
		},
		{
			"success, owner not exist",
			func(require *require.Assertions) {
				req = &nft.QueryNFTsOfClassRequest{
					ClassId: testClassID,
					Owner:   suite.addrs[1].String(),
				}
			},
			"",
			func(require *require.Assertions, res *nft.QueryNFTsOfClassResponse) {
				require.Len(res.Nfts, 0)
			},
		},
		{
			"Success, query by classId",
			func(require *require.Assertions) {
				req = &nft.QueryNFTsOfClassRequest{
					ClassId: testClassID,
				}
				nfts = []*nft.NFT{
					{
						ClassId: testClassID,
						Id:      testID,
						Uri:     testURI,
					},
				}
			},
			"",
			func(require *require.Assertions, res *nft.QueryNFTsOfClassResponse) {
				require.Equal(res.Nfts, nfts)
			},
		},
		{
			"Success,query by classId and owner",
			func(require *require.Assertions) {
				req = &nft.QueryNFTsOfClassRequest{
					ClassId: testClassID,
					Owner:   suite.addrs[0].String(),
				}
				nfts = []*nft.NFT{
					{
						ClassId: testClassID,
						Id:      testID,
						Uri:     testURI,
					},
				}
			},
			"",
			func(require *require.Assertions, res *nft.QueryNFTsOfClassResponse) {
				require.Equal(res.Nfts, nfts)
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

func (suite *TestSuite) TestGRPCQueryNFT() {
	var (
		req    *nft.QueryNFTRequest
		expNFT nft.NFT
	)
	testCases := []struct {
		msg      string
		malleate func(require *require.Assertions)
		expError string
		postTest func(require *require.Assertions, res *nft.QueryNFTResponse)
	}{
		{
			"fail empty ClassId",
			func(require *require.Assertions) {
				req = &nft.QueryNFTRequest{}
			},
			"class id can not be empty",
			func(require *require.Assertions, res *nft.QueryNFTResponse) {},
		},
		{
			"fail empty nft id",
			func(require *require.Assertions) {
				req = &nft.QueryNFTRequest{
					ClassId: testClassID,
				}
			},
			"nft id can not be empty",
			func(require *require.Assertions, res *nft.QueryNFTResponse) {},
		},
		{
			"fail ClassId not exist",
			func(require *require.Assertions) {
				req = &nft.QueryNFTRequest{
					ClassId: "kitty1",
					Id:      testID,
				}
				suite.TestMint()
			},
			"not found nft",
			func(require *require.Assertions, res *nft.QueryNFTResponse) {},
		},
		{
			"fail nft id not exist",
			func(require *require.Assertions) {
				req = &nft.QueryNFTRequest{
					ClassId: testClassID,
					Id:      "kitty2",
				}
			},
			"not found nft",
			func(require *require.Assertions, res *nft.QueryNFTResponse) {},
		},
		{
			"success",
			func(require *require.Assertions) {
				req = &nft.QueryNFTRequest{
					ClassId: testClassID,
					Id:      testID,
				}
				expNFT = nft.NFT{
					ClassId: testClassID,
					Id:      testID,
					Uri:     testURI,
				}
			},
			"",
			func(require *require.Assertions, res *nft.QueryNFTResponse) {
				require.Equal(*res.Nft, expNFT)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := suite.Require()
			tc.malleate(require)
			result, err := suite.queryClient.NFT(gocontext.Background(), req)
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

func (suite *TestSuite) TestGRPCQueryClass() {
	var (
		req   *nft.QueryClassRequest
		class nft.Class
	)
	testCases := []struct {
		msg      string
		malleate func(require *require.Assertions)
		expError string
		postTest func(require *require.Assertions, res *nft.QueryClassResponse)
	}{
		{
			"fail empty ClassId",
			func(require *require.Assertions) {
				req = &nft.QueryClassRequest{}
			},
			"class id can not be empty",
			func(require *require.Assertions, res *nft.QueryClassResponse) {},
		},
		{
			"fail ClassId not exist",
			func(require *require.Assertions) {
				req = &nft.QueryClassRequest{
					ClassId: "kitty1",
				}
				suite.TestSaveClass()
			},
			"not found class",
			func(require *require.Assertions, res *nft.QueryClassResponse) {},
		},
		{
			"success",
			func(require *require.Assertions) {
				class = nft.Class{
					Id:          testClassID,
					Name:        testClassName,
					Symbol:      testClassSymbol,
					Description: testClassDescription,
					Uri:         testClassURI,
					UriHash:     testClassURIHash,
				}
				req = &nft.QueryClassRequest{
					ClassId: testClassID,
				}
			},
			"",
			func(require *require.Assertions, res *nft.QueryClassResponse) {
				require.Equal(*res.Class, class)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := suite.Require()
			tc.malleate(require)
			result, err := suite.queryClient.Class(gocontext.Background(), req)
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

func (suite *TestSuite) TestGRPCQueryClasses() {
	var (
		req     *nft.QueryClassesRequest
		classes []nft.Class
	)
	testCases := []struct {
		msg      string
		malleate func(require *require.Assertions)
		expError string
		postTest func(require *require.Assertions, res *nft.QueryClassesResponse)
	}{
		{
			"success Class not exist",
			func(require *require.Assertions) {
				req = &nft.QueryClassesRequest{}
			},
			"",
			func(require *require.Assertions, res *nft.QueryClassesResponse) {
				require.Len(res.Classes, 0)
			},
		},
		{
			"success",
			func(require *require.Assertions) {
				req = &nft.QueryClassesRequest{}
				classes = []nft.Class{
					{
						Id:          testClassID,
						Name:        testClassName,
						Symbol:      testClassSymbol,
						Description: testClassDescription,
						Uri:         testClassURI,
						UriHash:     testClassURIHash,
					},
				}
				suite.TestSaveClass()
			},
			"",
			func(require *require.Assertions, res *nft.QueryClassesResponse) {
				require.Len(res.Classes, 1)
				require.Equal(*res.Classes[0], classes[0])
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := suite.Require()
			tc.malleate(require)
			result, err := suite.queryClient.Classes(gocontext.Background(), req)
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
