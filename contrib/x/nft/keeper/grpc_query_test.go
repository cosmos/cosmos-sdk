package keeper_test

import (
	gocontext "context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec/address"
	nft2 "github.com/cosmos/cosmos-sdk/contrib/x/nft"
)

func TestGRPCQuery(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestBalance() {
	s.accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()
	var req *nft2.QueryBalanceRequest
	testCases := []struct {
		msg      string
		malleate func(index int, require *require.Assertions)
		expError string
		balance  uint64
		postTest func(index int, require *require.Assertions, res *nft2.QueryBalanceResponse, expBalance uint64)
	}{
		{
			"fail empty ClassId",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryBalanceRequest{}
			},
			nft2.ErrEmptyClassID.Error(),
			0,
			func(index int, require *require.Assertions, res *nft2.QueryBalanceResponse, expBalance uint64) {},
		},
		{
			"fail invalid Owner addr",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryBalanceRequest{
					ClassId: testClassID,
					Owner:   "owner",
				}
			},
			"decoding bech32 failed",
			0,
			func(index int, require *require.Assertions, res *nft2.QueryBalanceResponse, expBalance uint64) {},
		},
		{
			"Success",
			func(index int, require *require.Assertions) {
				s.TestMint()
				req = &nft2.QueryBalanceRequest{
					ClassId: testClassID,
					Owner:   s.encodedAddrs[0],
				}
			},
			"",
			2,
			func(index int, require *require.Assertions, res *nft2.QueryBalanceResponse, expBalance uint64) {
				require.Equal(res.Amount, expBalance, "the error occurred on:%d", index)
			},
		},
	}
	for index, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := s.Require()
			tc.malleate(index, require)
			result, err := s.queryClient.Balance(gocontext.Background(), req)
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

func (s *TestSuite) TestOwner() {
	var (
		req   *nft2.QueryOwnerRequest
		owner string
	)
	testCases := []struct {
		msg      string
		malleate func(index int, require *require.Assertions)
		expError string
		postTest func(index int, require *require.Assertions, res *nft2.QueryOwnerResponse)
	}{
		{
			"fail empty ClassId",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryOwnerRequest{
					Id: testID,
				}
			},
			nft2.ErrEmptyClassID.Error(),
			func(index int, require *require.Assertions, res *nft2.QueryOwnerResponse) {},
		},
		{
			"fail empty nft id",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryOwnerRequest{
					ClassId: testClassID,
				}
			},
			nft2.ErrEmptyNFTID.Error(),
			func(index int, require *require.Assertions, res *nft2.QueryOwnerResponse) {},
		},
		{
			"success but nft id not exist",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryOwnerRequest{
					ClassId: testClassID,
					Id:      "kitty2",
				}
			},
			"",
			func(index int, require *require.Assertions, res *nft2.QueryOwnerResponse) {
				require.Equal(res.Owner, owner, "the error occurred on:%d", index)
			},
		},
		{
			"success but class id not exist",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryOwnerRequest{
					ClassId: "kitty1",
					Id:      testID,
				}
			},
			"",
			func(index int, require *require.Assertions, res *nft2.QueryOwnerResponse) {
				require.Equal(res.Owner, owner, "the error occurred on:%d", index)
			},
		},
		{
			"Success",
			func(index int, require *require.Assertions) {
				s.TestMint()
				req = &nft2.QueryOwnerRequest{
					ClassId: testClassID,
					Id:      testID,
				}
				owner = s.encodedAddrs[0]
			},
			"",
			func(index int, require *require.Assertions, res *nft2.QueryOwnerResponse) {
				require.Equal(res.Owner, owner, "the error occurred on:%d", index)
			},
		},
	}
	for index, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := s.Require()
			tc.malleate(index, require)
			result, err := s.queryClient.Owner(gocontext.Background(), req)
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

func (s *TestSuite) TestSupply() {
	var req *nft2.QuerySupplyRequest
	testCases := []struct {
		msg      string
		malleate func(index int, require *require.Assertions)
		expError string
		supply   uint64
		postTest func(index int, require *require.Assertions, res *nft2.QuerySupplyResponse, supply uint64)
	}{
		{
			"fail empty ClassId",
			func(index int, require *require.Assertions) {
				req = &nft2.QuerySupplyRequest{}
			},
			nft2.ErrEmptyClassID.Error(),
			0,
			func(index int, require *require.Assertions, res *nft2.QuerySupplyResponse, supply uint64) {},
		},
		{
			"success but class id not exist",
			func(index int, require *require.Assertions) {
				req = &nft2.QuerySupplyRequest{
					ClassId: "kitty1",
				}
			},
			"",
			0,
			func(index int, require *require.Assertions, res *nft2.QuerySupplyResponse, supply uint64) {
				require.Equal(res.Amount, supply, "the error occurred on:%d", index)
			},
		},
		{
			"success but supply equal zero",
			func(index int, require *require.Assertions) {
				req = &nft2.QuerySupplyRequest{
					ClassId: testClassID,
				}
				s.TestSaveClass()
			},
			"",
			0,
			func(index int, require *require.Assertions, res *nft2.QuerySupplyResponse, supply uint64) {
				require.Equal(res.Amount, supply, "the error occurred on:%d", index)
			},
		},
		{
			"Success",
			func(index int, require *require.Assertions) {
				n := nft2.NFT{
					ClassId: testClassID,
					Id:      testID,
					Uri:     testURI,
				}
				err := s.nftKeeper.Mint(s.ctx, n, s.addrs[0])
				require.NoError(err, "the error occurred on:%d", index)

				req = &nft2.QuerySupplyRequest{
					ClassId: testClassID,
				}
			},
			"",
			1,
			func(index int, require *require.Assertions, res *nft2.QuerySupplyResponse, supply uint64) {
				require.Equal(res.Amount, supply, "the error occurred on:%d", index)
			},
		},
	}
	for index, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := s.Require()
			tc.malleate(index, require)
			result, err := s.queryClient.Supply(gocontext.Background(), req)
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

func (s *TestSuite) TestNFTs() {
	var (
		req  *nft2.QueryNFTsRequest
		nfts []*nft2.NFT
	)
	testCases := []struct {
		msg      string
		malleate func(index int, require *require.Assertions)
		expError string
		postTest func(index int, require *require.Assertions, res *nft2.QueryNFTsResponse)
	}{
		{
			"fail empty Owner and ClassId",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryNFTsRequest{}
			},
			"must provide at least one of classID or owner",
			func(index int, require *require.Assertions, res *nft2.QueryNFTsResponse) {},
		},
		{
			"success,empty ClassId and no nft",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryNFTsRequest{
					Owner: s.encodedAddrs[1],
				}
				s.TestSaveClass()
			},
			"",
			func(index int, require *require.Assertions, res *nft2.QueryNFTsResponse) {
				require.Len(res.Nfts, 0, "the error occurred on:%d", index)
			},
		},
		{
			"success, empty Owner and class id not exist",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryNFTsRequest{
					ClassId: "kitty1",
				}
				n := nft2.NFT{
					ClassId: testClassID,
					Id:      testID,
					Uri:     testURI,
				}
				err := s.nftKeeper.Mint(s.ctx, n, s.addrs[0])
				require.NoError(err, "the error occurred on:%d", index)
			},
			"",
			func(index int, require *require.Assertions, res *nft2.QueryNFTsResponse) {
				require.Len(res.Nfts, 0, "the error occurred on:%d", index)
			},
		},
		{
			"Success,query by owner",
			func(index int, require *require.Assertions) {
				err := s.nftKeeper.SaveClass(s.ctx, nft2.Class{
					Id: "MyKitty",
				})
				require.NoError(err)

				nfts = []*nft2.NFT{}
				for i := 0; i < 5; i++ {
					n := nft2.NFT{
						ClassId: "MyKitty",
						Id:      fmt.Sprintf("MyCat%d", i),
					}
					err := s.nftKeeper.Mint(s.ctx, n, s.addrs[2])
					require.NoError(err)
					nfts = append(nfts, &n)
				}

				req = &nft2.QueryNFTsRequest{
					Owner: s.encodedAddrs[2],
				}
			},
			"",
			func(index int, require *require.Assertions, res *nft2.QueryNFTsResponse) {
				require.EqualValues(res.Nfts, nfts, "the error occurred on:%d", index)
			},
		},
		{
			"Success,query by classID",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryNFTsRequest{
					ClassId: "MyKitty",
				}
			},
			"",
			func(index int, require *require.Assertions, res *nft2.QueryNFTsResponse) {
				require.EqualValues(res.Nfts, nfts, "the error occurred on:%d", index)
			},
		},
		{
			"Success,query by classId and owner",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryNFTsRequest{
					ClassId: testClassID,
					Owner:   s.encodedAddrs[0],
				}
				nfts = []*nft2.NFT{
					{
						ClassId: testClassID,
						Id:      testID,
						Uri:     testURI,
					},
				}
			},
			"",
			func(index int, require *require.Assertions, res *nft2.QueryNFTsResponse) {
				require.Equal(res.Nfts, nfts, "the error occurred on:%d", index)
			},
		},
	}
	for index, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := s.Require()
			tc.malleate(index, require)
			result, err := s.queryClient.NFTs(gocontext.Background(), req)
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

func (s *TestSuite) TestNFT() {
	var (
		req    *nft2.QueryNFTRequest
		expNFT nft2.NFT
	)
	testCases := []struct {
		msg      string
		malleate func(index int, require *require.Assertions)
		expError string
		postTest func(index int, require *require.Assertions, res *nft2.QueryNFTResponse)
	}{
		{
			"fail empty ClassId",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryNFTRequest{}
			},
			nft2.ErrEmptyClassID.Error(),
			func(index int, require *require.Assertions, res *nft2.QueryNFTResponse) {},
		},
		{
			"fail empty nft id",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryNFTRequest{
					ClassId: testClassID,
				}
			},
			nft2.ErrEmptyNFTID.Error(),
			func(index int, require *require.Assertions, res *nft2.QueryNFTResponse) {},
		},
		{
			"fail ClassId not exist",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryNFTRequest{
					ClassId: "kitty1",
					Id:      testID,
				}
				s.TestMint()
			},
			"not found nft",
			func(index int, require *require.Assertions, res *nft2.QueryNFTResponse) {},
		},
		{
			"fail nft id not exist",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryNFTRequest{
					ClassId: testClassID,
					Id:      "kitty2",
				}
			},
			"not found nft",
			func(index int, require *require.Assertions, res *nft2.QueryNFTResponse) {},
		},
		{
			"success",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryNFTRequest{
					ClassId: testClassID,
					Id:      testID,
				}
				expNFT = nft2.NFT{
					ClassId: testClassID,
					Id:      testID,
					Uri:     testURI,
				}
			},
			"",
			func(index int, require *require.Assertions, res *nft2.QueryNFTResponse) {
				require.Equal(*res.Nft, expNFT, "the error occurred on:%d", index)
			},
		},
	}
	for index, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := s.Require()
			tc.malleate(index, require)
			result, err := s.queryClient.NFT(gocontext.Background(), req)
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

func (s *TestSuite) TestClass() {
	var (
		req   *nft2.QueryClassRequest
		class nft2.Class
	)
	testCases := []struct {
		msg      string
		malleate func(index int, require *require.Assertions)
		expError string
		postTest func(index int, require *require.Assertions, res *nft2.QueryClassResponse)
	}{
		{
			"fail empty ClassId",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryClassRequest{}
			},
			nft2.ErrEmptyClassID.Error(),
			func(index int, require *require.Assertions, res *nft2.QueryClassResponse) {},
		},
		{
			"fail ClassId not exist",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryClassRequest{
					ClassId: "kitty1",
				}
				s.TestSaveClass()
			},
			"not found class",
			func(index int, require *require.Assertions, res *nft2.QueryClassResponse) {},
		},
		{
			"success",
			func(index int, require *require.Assertions) {
				class = nft2.Class{
					Id:          testClassID,
					Name:        testClassName,
					Symbol:      testClassSymbol,
					Description: testClassDescription,
					Uri:         testClassURI,
					UriHash:     testClassURIHash,
				}
				req = &nft2.QueryClassRequest{
					ClassId: testClassID,
				}
			},
			"",
			func(index int, require *require.Assertions, res *nft2.QueryClassResponse) {
				require.Equal(*res.Class, class, "the error occurred on:%d", index)
			},
		},
	}
	for index, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := s.Require()
			tc.malleate(index, require)
			result, err := s.queryClient.Class(gocontext.Background(), req)
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

func (s *TestSuite) TestClasses() {
	var (
		req     *nft2.QueryClassesRequest
		classes []nft2.Class
	)
	testCases := []struct {
		msg      string
		malleate func(index int, require *require.Assertions)
		expError string
		postTest func(index int, require *require.Assertions, res *nft2.QueryClassesResponse)
	}{
		{
			"success Class not exist",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryClassesRequest{}
			},
			"",
			func(index int, require *require.Assertions, res *nft2.QueryClassesResponse) {
				require.Len(res.Classes, 0)
			},
		},
		{
			"success",
			func(index int, require *require.Assertions) {
				req = &nft2.QueryClassesRequest{}
				classes = []nft2.Class{
					{
						Id:          testClassID,
						Name:        testClassName,
						Symbol:      testClassSymbol,
						Description: testClassDescription,
						Uri:         testClassURI,
						UriHash:     testClassURIHash,
					},
				}
				s.TestSaveClass()
			},
			"",
			func(index int, require *require.Assertions, res *nft2.QueryClassesResponse) {
				require.Len(res.Classes, 1, "the error occurred on:%d", index)
				require.Equal(*res.Classes[0], classes[0], "the error occurred on:%d", index)
			},
		},
	}
	for index, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := s.Require()
			tc.malleate(index, require)
			result, err := s.queryClient.Classes(gocontext.Background(), req)
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
