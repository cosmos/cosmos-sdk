package testutil

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/testutil/rest"
	"github.com/cosmos/cosmos-sdk/x/nft"
)

func (s *IntegrationTestSuite) TestQueryBalanceGRPC() {
	val := s.network.Validators[0]
	testCases := []struct {
		name string
		args struct {
			ClassId string
			Owner   string
		}
		expectErr   bool
		errMsg      string
		expectValue uint64
	}{
		{
			name: "fail not exist class id",
			args: struct {
				ClassId string
				Owner   string
			}{
				ClassId: "invalid_class_id",
				Owner:   s.owner.String(),
			},
			expectErr:   true,
			errMsg:      "invalid class id",
			expectValue: 0,
		},
		{
			name: "fail not exist owner",
			args: struct {
				ClassId string
				Owner   string
			}{
				ClassId: ExpNFT.ClassId,
				Owner:   s.owner.String(),
			},
			expectErr:   false,
			expectValue: 0,
		},
		{
			name: "success",
			args: struct {
				ClassId string
				Owner   string
			}{
				ClassId: ExpNFT.ClassId,
				Owner:   val.Address.String(),
			},
			expectErr:   false,
			expectValue: 1,
		},
	}
	balanceURL := val.APIAddress + "/cosmos/nft/v1beta1/balance/%s/%s"
	for _, tc := range testCases {
		uri := fmt.Sprintf(balanceURL, tc.args.Owner, tc.args.ClassId)
		s.Run(tc.name, func() {
			resp, _ := rest.GetRequest(uri)
			if tc.expectErr {
				s.Require().Contains(string(resp), tc.errMsg)
			} else {
				var g nft.QueryBalanceResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(resp, &g)
				s.Require().NoError(err)
				s.Require().Equal(tc.expectValue, g.Amount)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryOwnerGRPC() {
	val := s.network.Validators[0]

	testCases := []struct {
		name string
		args struct {
			ClassId string
			Id      string
		}
		expectErr    bool
		errMsg       string
		expectResult string
	}{
		{
			name: "class id is invalid",
			args: struct {
				ClassId string
				Id      string
			}{
				ClassId: "invalid_class_id",
				Id:      ExpNFT.Id,
			},
			expectErr:    true,
			errMsg:       "invalid class id",
			expectResult: "",
		},
		{
			name: "class id does not exist",
			args: struct {
				ClassId string
				Id      string
			}{
				ClassId: "class-id",
				Id:      ExpNFT.Id,
			},
			expectErr:    false,
			expectResult: "",
		},
		{
			name: "nft id is invalid",
			args: struct {
				ClassId string
				Id      string
			}{
				ClassId: ExpNFT.ClassId,
				Id:      "invalid_nft_id",
			},
			expectErr:    true,
			expectResult: "",
		},
		{
			name: "nft id does not exist",
			args: struct {
				ClassId string
				Id      string
			}{
				ClassId: ExpNFT.ClassId,
				Id:      "nft-id",
			},
			expectErr:    false,
			expectResult: "",
		},
		{
			name: "nft exist",
			args: struct {
				ClassId string
				Id      string
			}{
				ClassId: ExpNFT.ClassId,
				Id:      ExpNFT.Id,
			},
			expectErr:    false,
			expectResult: val.Address.String(),
		},
	}
	ownerURL := val.APIAddress + "/cosmos/nft/v1beta1/owner/%s/%s"
	for _, tc := range testCases {
		uri := fmt.Sprintf(ownerURL, tc.args.ClassId, tc.args.Id)
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(uri)
			if tc.expectErr {
				s.Require().Contains(string(resp), tc.errMsg)
			} else {
				s.Require().NoError(err)
				var result nft.QueryOwnerResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp, &result)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQuerySupplyGRPC() {
	val := s.network.Validators[0]

	testCases := []struct {
		name string
		args struct {
			ClassId string
		}
		expectErr    bool
		errMsg       string
		expectResult uint64
	}{
		{
			name: "class id is invalid",
			args: struct {
				ClassId string
			}{
				ClassId: "invalid_class_id",
			},
			expectErr:    true,
			errMsg:       "invalid class id",
			expectResult: 0,
		},
		{
			name: "class id does not exist",
			args: struct {
				ClassId string
			}{
				ClassId: "class-id",
			},
			expectErr:    false,
			expectResult: 0,
		},
		{
			name: "class id exist",
			args: struct {
				ClassId string
			}{
				ClassId: ExpNFT.ClassId,
			},
			expectErr:    false,
			expectResult: 1,
		},
	}
	supplyURL := val.APIAddress + "/cosmos/nft/v1beta1/supply/%s"
	for _, tc := range testCases {
		uri := fmt.Sprintf(supplyURL, tc.args.ClassId)
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(uri)
			if tc.expectErr {
				s.Require().Contains(string(resp), tc.errMsg)
			} else {
				s.Require().NoError(err)
				var result nft.QuerySupplyResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp, &result)
				s.Require().NoError(err)
				s.Require().EqualValues(tc.expectResult, result.Amount)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryNFTsByOwnerGRPC() {
	val := s.network.Validators[0]
	testCases := []struct {
		name string
		args struct {
			ClassId string
			Owner   string
		}
		expectErr    bool
		errorMsg     string
		expectResult []*nft.NFT
	}{
		{
			name: "class id is invalid",
			args: struct {
				ClassId string
				Owner   string
			}{
				ClassId: "invalid_class_id",
				Owner:   s.owner.String(),
			},
			expectErr:    true,
			errorMsg:     "invalid class id",
			expectResult: []*nft.NFT{},
		},
		{
			name: "class id does not exist",
			args: struct {
				ClassId string
				Owner   string
			}{
				ClassId: "class-id",
				Owner:   s.owner.String(),
			},
			expectErr:    false,
			expectResult: []*nft.NFT{},
		},
		{
			name: "owner does not exist",
			args: struct {
				ClassId string
				Owner   string
			}{
				ClassId: ExpNFT.ClassId,
				Owner:   s.owner.String(),
			},
			expectErr:    false,
			expectResult: []*nft.NFT{},
		},
		{
			name: "nft exist",
			args: struct {
				ClassId string
				Owner   string
			}{
				ClassId: ExpNFT.ClassId,
				Owner:   val.Address.String(),
			},
			expectErr:    false,
			expectResult: []*nft.NFT{&ExpNFT},
		},
	}
	nftsOfClassURL := val.APIAddress + "/cosmos/nft/v1beta1/nfts/%s?owner=%s"
	for _, tc := range testCases {
		uri := fmt.Sprintf(nftsOfClassURL, tc.args.ClassId, tc.args.Owner)
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(uri)
			if tc.expectErr {
				s.Require().Contains(string(resp), tc.errorMsg)
			} else {
				s.Require().NoError(err)
				var result nft.QueryNFTsOfClassResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp, &result)
				s.Require().NoError(err)
				s.Require().EqualValues(tc.expectResult, result.Nfts)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryNFTsOfClassGRPC() {
	val := s.network.Validators[0]
	testCases := []struct {
		name string
		args struct {
			ClassId string
		}
		expectErr    bool
		errorMsg     string
		expectResult []*nft.NFT
	}{
		{
			name: "class id is invalid",
			args: struct {
				ClassId string
			}{
				ClassId: "invalid_class_id",
			},
			expectErr:    true,
			expectResult: []*nft.NFT{},
		},
		{
			name: "class id does not exist",
			args: struct {
				ClassId string
			}{
				ClassId: "class-id",
			},
			expectErr:    false,
			expectResult: []*nft.NFT{},
		},
		{
			name: "class id exist",
			args: struct {
				ClassId string
			}{
				ClassId: ExpNFT.ClassId,
			},
			expectErr:    false,
			expectResult: []*nft.NFT{&ExpNFT},
		},
	}
	nftsOfClassURL := val.APIAddress + "/cosmos/nft/v1beta1/nfts/%s"
	for _, tc := range testCases {
		uri := fmt.Sprintf(nftsOfClassURL, tc.args.ClassId)
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(uri)
			if tc.expectErr {
				s.Require().Contains(string(resp), tc.errorMsg)
			} else {
				s.Require().NoError(err)
				var result nft.QueryNFTsOfClassResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp, &result)
				s.Require().NoError(err)
				s.Require().EqualValues(tc.expectResult, result.Nfts)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryNFTGRPC() {
	val := s.network.Validators[0]
	testCases := []struct {
		name string
		args struct {
			ClassId string
			Id      string
		}
		expectErr bool
		errorMsg  string
	}{
		{
			name: "class id is invalid",
			args: struct {
				ClassId string
				Id      string
			}{
				ClassId: "invalid_class_id",
				Id:      ExpNFT.Id,
			},
			expectErr: true,
			errorMsg:  "invalid class id",
		},
		{
			name: "class id does not exist",
			args: struct {
				ClassId string
				Id      string
			}{
				ClassId: "class",
				Id:      ExpNFT.Id,
			},
			expectErr: true,
			errorMsg:  "not found nft",
		},
		{
			name: "nft id is invalid",
			args: struct {
				ClassId string
				Id      string
			}{
				ClassId: ExpNFT.ClassId,
				Id:      "invalid_nft_id",
			},
			expectErr: true,
			errorMsg:  "invalid nft id",
		},
		{
			name: "nft id does not exist",
			args: struct {
				ClassId string
				Id      string
			}{
				ClassId: ExpNFT.ClassId,
				Id:      "nft-id",
			},
			expectErr: true,
			errorMsg:  "not found nft",
		},
		{
			name: "exist nft",
			args: struct {
				ClassId string
				Id      string
			}{
				ClassId: ExpNFT.ClassId,
				Id:      ExpNFT.Id,
			},
			expectErr: false,
		},
	}
	nftURL := val.APIAddress + "/cosmos/nft/v1beta1/nfts/%s/%s"
	for _, tc := range testCases {
		uri := fmt.Sprintf(nftURL, tc.args.ClassId, tc.args.Id)
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(uri)
			if tc.expectErr {
				s.Require().Contains(string(resp), tc.errorMsg)
			} else {
				s.Require().NoError(err)
				var result nft.QueryNFTResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp, &result)
				s.Require().NoError(err)
				s.Require().EqualValues(ExpNFT, *result.Nft)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryClassGRPC() {
	val := s.network.Validators[0]
	testCases := []struct {
		name string
		args struct {
			ClassId string
		}
		expectErr bool
		errorMsg  string
	}{
		{
			name: "class id does not exist",
			args: struct {
				ClassId string
			}{
				ClassId: "class-id",
			},
			expectErr: true,
			errorMsg:  "not found class",
		},
		{
			name: "class id exist",
			args: struct {
				ClassId string
			}{
				ClassId: ExpNFT.ClassId,
			},
			expectErr: false,
		},
	}
	classURL := val.APIAddress + "/cosmos/nft/v1beta1/classes/%s"
	for _, tc := range testCases {
		uri := fmt.Sprintf(classURL, tc.args.ClassId)
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(uri)
			if tc.expectErr {
				s.Require().Contains(string(resp), tc.errorMsg)
			} else {
				s.Require().NoError(err)
				var result nft.QueryClassResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp, &result)
				s.Require().NoError(err)
				s.Require().EqualValues(ExpClass, *result.Class)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryClassesGRPC() {
	val := s.network.Validators[0]
	classURL := val.APIAddress + "/cosmos/nft/v1beta1/classes"
	resp, err := rest.GetRequest(classURL)
	s.Require().NoError(err)
	var result nft.QueryClassesResponse
	err = val.ClientCtx.Codec.UnmarshalJSON(resp, &result)
	s.Require().NoError(err)
	s.Require().Len(result.Classes, 1)
	s.Require().EqualValues(ExpClass, *result.Classes[0])
}
