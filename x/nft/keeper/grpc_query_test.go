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

func (suite *TestSuite) TestBalance() {
	var (
		req *nft.QueryBalanceRequest
	)
	testCases := []struct {
		msg      string
		malleate func(index int, require *require.Assertions)
		expError string
		balance  uint64
		postTest func(index int, require *require.Assertions, res *nft.QueryBalanceResponse, expBalance uint64)
	}{
		{
			"fail empty ClassId",
			func(index int, require *require.Assertions) {
				req = &nft.QueryBalanceRequest{}
			},
			"invalid class id",
			0,
			func(index int, require *require.Assertions, res *nft.QueryBalanceResponse, expBalance uint64) {},
		},
		{
			"fail invalid Owner addr",
			func(index int, require *require.Assertions) {
				req = &nft.QueryBalanceRequest{
					ClassId: testClassID,
					Owner:   "owner",
				}
			},
			"decoding bech32 failed",
			0,
			func(index int, require *require.Assertions, res *nft.QueryBalanceResponse, expBalance uint64) {},
		},
		{
			"Success",
			func(index int, require *require.Assertions) {
				suite.TestMint()
				req = &nft.QueryBalanceRequest{
					ClassId: testClassID,
					Owner:   suite.addrs[0].String(),
				}
			},
			"",
			1,
			func(index int, require *require.Assertions, res *nft.QueryBalanceResponse, expBalance uint64) {
				require.Equal(res.Amount, expBalance, "the error occurred on:%d", index)
			},
		},
	}
	for index, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := suite.Require()
			tc.malleate(index, require)
			result, err := suite.queryClient.Balance(gocontext.Background(), req)
			if tc.expError == "" {
				require.NoError(err)
			} else {
				require.Error(err)
				require.Contains(err.Error(), tc.expError)
			}
			tc.postTest(index, require, result, tc.balance)
		})
	}
}

func (suite *TestSuite) TestOwner() {
	var (
		req   *nft.QueryOwnerRequest
		owner string
	)
	testCases := []struct {
		msg      string
		malleate func(index int, require *require.Assertions)
		expError string
		postTest func(index int, require *require.Assertions, res *nft.QueryOwnerResponse)
	}{
		{
			"fail empty ClassId",
			func(index int, require *require.Assertions) {
				req = &nft.QueryOwnerRequest{
					Id: testID,
				}
			},
			"invalid class id",
			func(index int, require *require.Assertions, res *nft.QueryOwnerResponse) {},
		},
		{
			"fail empty nft id",
			func(index int, require *require.Assertions) {
				req = &nft.QueryOwnerRequest{
					ClassId: testClassID,
				}
			},
			"invalid nft id",
			func(index int, require *require.Assertions, res *nft.QueryOwnerResponse) {},
		},
		{
			"success but nft id not exist",
			func(index int, require *require.Assertions) {
				req = &nft.QueryOwnerRequest{
					ClassId: testClassID,
					Id:      "kitty2",
				}
			},
			"",
			func(index int, require *require.Assertions, res *nft.QueryOwnerResponse) {
				require.Equal(res.Owner, owner, "the error occurred on:%d", index)
			},
		},
		{
			"success but class id not exist",
			func(index int, require *require.Assertions) {
				req = &nft.QueryOwnerRequest{
					ClassId: "kitty1",
					Id:      testID,
				}
			},
			"",
			func(index int, require *require.Assertions, res *nft.QueryOwnerResponse) {
				require.Equal(res.Owner, owner, "the error occurred on:%d", index)
			},
		},
		{
			"Success",
			func(index int, require *require.Assertions) {
				suite.TestMint()
				req = &nft.QueryOwnerRequest{
					ClassId: testClassID,
					Id:      testID,
				}
				owner = suite.addrs[0].String()
			},
			"",
			func(index int, require *require.Assertions, res *nft.QueryOwnerResponse) {
				require.Equal(res.Owner, owner, "the error occurred on:%d", index)
			},
		},
	}
	for index, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := suite.Require()
			tc.malleate(index, require)
			result, err := suite.queryClient.Owner(gocontext.Background(), req)
			if tc.expError == "" {
				require.NoError(err)
			} else {
				require.Error(err)
				require.Contains(err.Error(), tc.expError)
			}
			tc.postTest(index, require, result)
		})
	}
}

func (suite *TestSuite) TestSupply() {
	var (
		req *nft.QuerySupplyRequest
	)
	testCases := []struct {
		msg      string
		malleate func(index int, require *require.Assertions)
		expError string
		supply   uint64
		postTest func(index int, require *require.Assertions, res *nft.QuerySupplyResponse, supply uint64)
	}{
		{
			"fail empty ClassId",
			func(index int, require *require.Assertions) {
				req = &nft.QuerySupplyRequest{}
			},
			"invalid class id",
			0,
			func(index int, require *require.Assertions, res *nft.QuerySupplyResponse, supply uint64) {},
		},
		{
			"success but class id not exist",
			func(index int, require *require.Assertions) {
				req = &nft.QuerySupplyRequest{
					ClassId: "kitty1",
				}
			},
			"",
			0,
			func(index int, require *require.Assertions, res *nft.QuerySupplyResponse, supply uint64) {
				require.Equal(res.Amount, supply, "the error occurred on:%d", index)
			},
		},
		{
			"success but supply equal zero",
			func(index int, require *require.Assertions) {
				req = &nft.QuerySupplyRequest{
					ClassId: testClassID,
				}
				suite.TestSaveClass()
			},
			"",
			0,
			func(index int, require *require.Assertions, res *nft.QuerySupplyResponse, supply uint64) {
				require.Equal(res.Amount, supply, "the error occurred on:%d", index)
			},
		},
		{
			"Success",
			func(index int, require *require.Assertions) {
				n := nft.NFT{
					ClassId: testClassID,
					Id:      testID,
					Uri:     testURI,
				}
				err := suite.app.NFTKeeper.Mint(suite.ctx, n, suite.addrs[0])
				require.NoError(err, "the error occurred on:%d", index)

				req = &nft.QuerySupplyRequest{
					ClassId: testClassID,
				}
			},
			"",
			1,
			func(index int, require *require.Assertions, res *nft.QuerySupplyResponse, supply uint64) {
				require.Equal(res.Amount, supply, "the error occurred on:%d", index)
			},
		},
	}
	for index, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := suite.Require()
			tc.malleate(index, require)
			result, err := suite.queryClient.Supply(gocontext.Background(), req)
			if tc.expError == "" {
				require.NoError(err)
			} else {
				require.Error(err)
				require.Contains(err.Error(), tc.expError)
			}
			tc.postTest(index, require, result, tc.supply)
		})
	}
}

func (suite *TestSuite) TestNFTsOfClass() {
	var (
		req  *nft.QueryNFTsOfClassRequest
		nfts []*nft.NFT
	)
	testCases := []struct {
		msg      string
		malleate func(index int, require *require.Assertions)
		expError string
		postTest func(index int, require *require.Assertions, res *nft.QueryNFTsOfClassResponse)
	}{
		{
			"fail empty ClassId",
			func(index int, require *require.Assertions) {
				req = &nft.QueryNFTsOfClassRequest{}
			},
			"invalid class id",
			func(index int, require *require.Assertions, res *nft.QueryNFTsOfClassResponse) {},
		},
		{
			"success, no nft",
			func(index int, require *require.Assertions) {
				req = &nft.QueryNFTsOfClassRequest{
					ClassId: testClassID,
				}
				suite.TestSaveClass()
			},
			"",
			func(index int, require *require.Assertions, res *nft.QueryNFTsOfClassResponse) {
				require.Len(res.Nfts, 0, "the error occurred on:%d", index)
			},
		},
		{
			"success, class id not exist",
			func(index int, require *require.Assertions) {
				req = &nft.QueryNFTsOfClassRequest{
					ClassId: "kitty1",
				}
				n := nft.NFT{
					ClassId: testClassID,
					Id:      testID,
					Uri:     testURI,
				}
				err := suite.app.NFTKeeper.Mint(suite.ctx, n, suite.addrs[0])
				require.NoError(err, "the error occurred on:%d", index)
			},
			"",
			func(index int, require *require.Assertions, res *nft.QueryNFTsOfClassResponse) {
				require.Len(res.Nfts, 0, "the error occurred on:%d", index)
			},
		},
		{
			"success, owner not exist",
			func(index int, require *require.Assertions) {
				req = &nft.QueryNFTsOfClassRequest{
					ClassId: testClassID,
					Owner:   suite.addrs[1].String(),
				}
			},
			"",
			func(index int, require *require.Assertions, res *nft.QueryNFTsOfClassResponse) {
				require.Len(res.Nfts, 0, "the error occurred on:%d", index)
			},
		},
		{
			"Success, query by classId",
			func(index int, require *require.Assertions) {
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
			func(index int, require *require.Assertions, res *nft.QueryNFTsOfClassResponse) {
				require.Equal(res.Nfts, nfts, "the error occurred on:%d", index)
			},
		},
		{
			"Success,query by classId and owner",
			func(index int, require *require.Assertions) {
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
			func(index int, require *require.Assertions, res *nft.QueryNFTsOfClassResponse) {
				require.Equal(res.Nfts, nfts, "the error occurred on:%d", index)
			},
		},
	}
	for index, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := suite.Require()
			tc.malleate(index, require)
			result, err := suite.queryClient.NFTsOfClass(gocontext.Background(), req)
			if tc.expError == "" {
				require.NoError(err)
			} else {
				require.Error(err)
				require.Contains(err.Error(), tc.expError)
			}
			tc.postTest(index, require, result)
		})
	}
}

func (suite *TestSuite) TestNFT() {
	var (
		req    *nft.QueryNFTRequest
		expNFT nft.NFT
	)
	testCases := []struct {
		msg      string
		malleate func(index int, require *require.Assertions)
		expError string
		postTest func(index int, require *require.Assertions, res *nft.QueryNFTResponse)
	}{
		{
			"fail empty ClassId",
			func(index int, require *require.Assertions) {
				req = &nft.QueryNFTRequest{}
			},
			"invalid class id",
			func(index int, require *require.Assertions, res *nft.QueryNFTResponse) {},
		},
		{
			"fail empty nft id",
			func(index int, require *require.Assertions) {
				req = &nft.QueryNFTRequest{
					ClassId: testClassID,
				}
			},
			"invalid nft id",
			func(index int, require *require.Assertions, res *nft.QueryNFTResponse) {},
		},
		{
			"fail ClassId not exist",
			func(index int, require *require.Assertions) {
				req = &nft.QueryNFTRequest{
					ClassId: "kitty1",
					Id:      testID,
				}
				suite.TestMint()
			},
			"not found nft",
			func(index int, require *require.Assertions, res *nft.QueryNFTResponse) {},
		},
		{
			"fail nft id not exist",
			func(index int, require *require.Assertions) {
				req = &nft.QueryNFTRequest{
					ClassId: testClassID,
					Id:      "kitty2",
				}
			},
			"not found nft",
			func(index int, require *require.Assertions, res *nft.QueryNFTResponse) {},
		},
		{
			"success",
			func(index int, require *require.Assertions) {
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
			func(index int, require *require.Assertions, res *nft.QueryNFTResponse) {
				require.Equal(*res.Nft, expNFT, "the error occurred on:%d", index)
			},
		},
	}
	for index, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := suite.Require()
			tc.malleate(index, require)
			result, err := suite.queryClient.NFT(gocontext.Background(), req)
			if tc.expError == "" {
				require.NoError(err)
			} else {
				require.Error(err)
				require.Contains(err.Error(), tc.expError)
			}
			tc.postTest(index, require, result)
		})
	}
}

func (suite *TestSuite) TestClass() {
	var (
		req   *nft.QueryClassRequest
		class nft.Class
	)
	testCases := []struct {
		msg      string
		malleate func(index int, require *require.Assertions)
		expError string
		postTest func(index int, require *require.Assertions, res *nft.QueryClassResponse)
	}{
		{
			"fail empty ClassId",
			func(index int, require *require.Assertions) {
				req = &nft.QueryClassRequest{}
			},
			"invalid class id",
			func(index int, require *require.Assertions, res *nft.QueryClassResponse) {},
		},
		{
			"fail ClassId not exist",
			func(index int, require *require.Assertions) {
				req = &nft.QueryClassRequest{
					ClassId: "kitty1",
				}
				suite.TestSaveClass()
			},
			"not found class",
			func(index int, require *require.Assertions, res *nft.QueryClassResponse) {},
		},
		{
			"success",
			func(index int, require *require.Assertions) {
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
			func(index int, require *require.Assertions, res *nft.QueryClassResponse) {
				require.Equal(*res.Class, class, "the error occurred on:%d", index)
			},
		},
	}
	for index, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := suite.Require()
			tc.malleate(index, require)
			result, err := suite.queryClient.Class(gocontext.Background(), req)
			if tc.expError == "" {
				require.NoError(err)
			} else {
				require.Error(err)
				require.Contains(err.Error(), tc.expError)
			}
			tc.postTest(index, require, result)
		})
	}
}

func (suite *TestSuite) TestClasses() {
	var (
		req     *nft.QueryClassesRequest
		classes []nft.Class
	)
	testCases := []struct {
		msg      string
		malleate func(index int, require *require.Assertions)
		expError string
		postTest func(index int, require *require.Assertions, res *nft.QueryClassesResponse)
	}{
		{
			"success Class not exist",
			func(index int, require *require.Assertions) {
				req = &nft.QueryClassesRequest{}
			},
			"",
			func(index int, require *require.Assertions, res *nft.QueryClassesResponse) {
				require.Len(res.Classes, 0)
			},
		},
		{
			"success",
			func(index int, require *require.Assertions) {
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
			func(index int, require *require.Assertions, res *nft.QueryClassesResponse) {
				require.Len(res.Classes, 1, "the error occurred on:%d", index)
				require.Equal(*res.Classes[0], classes[0], "the error occurred on:%d", index)
			},
		},
	}
	for index, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := suite.Require()
			tc.malleate(index, require)
			result, err := suite.queryClient.Classes(gocontext.Background(), req)
			if tc.expError == "" {
				require.NoError(err)
			} else {
				require.Error(err)
				require.Contains(err.Error(), tc.expError)
			}
			tc.postTest(index, require, result)
		})
	}
}
