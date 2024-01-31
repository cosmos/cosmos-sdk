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
			ClassID string
			Owner   string
		}
		expectErr   bool
		errMsg      string
		expectValue uint64
	}{
		{
			name: "fail not exist class id",
			args: struct {
				ClassID string
				Owner   string
			}{
				ClassID: "invalid_class_id",
				Owner:   s.owner.String(),
			},
			expectErr:   true,
			errMsg:      "invalid class id",
			expectValue: 0,
		},
		{
			name: "fail not exist owner",
			args: struct {
				ClassID string
				Owner   string
			}{
				ClassID: ExpNFT.ClassId,
				Owner:   s.owner.String(),
			},
			expectErr:   false,
			expectValue: 0,
		},
		{
			name: "success",
			args: struct {
				ClassID string
				Owner   string
			}{
				ClassID: ExpNFT.ClassId,
				Owner:   val.Address.String(),
			},
			expectErr:   false,
			expectValue: 1,
		},
	}
	balanceURL := val.APIAddress + "/cosmos/nft/v1beta1/balance/%s/%s"
	for _, tc := range testCases {
		uri := fmt.Sprintf(balanceURL, tc.args.Owner, tc.args.ClassID)
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
			ClassID string
			ID      string
		}
		expectErr    bool
		errMsg       string
		expectResult string
	}{
		{
			name: "class id is invalid",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: "invalid_class_id",
				ID:      ExpNFT.Id,
			},
			expectErr:    true,
			errMsg:       "invalid class id",
			expectResult: "",
		},
		{
			name: "class id does not exist",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: "class-id",
				ID:      ExpNFT.Id,
			},
			expectErr:    false,
			expectResult: "",
		},
		{
			name: "nft id is invalid",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: ExpNFT.ClassId,
				ID:      "invalid_nft_id",
			},
			expectErr:    true,
			expectResult: "",
		},
		{
			name: "nft id does not exist",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: ExpNFT.ClassId,
				ID:      "nft-id",
			},
			expectErr:    false,
			expectResult: "",
		},
		{
			name: "nft exist",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: ExpNFT.ClassId,
				ID:      ExpNFT.Id,
			},
			expectErr:    false,
			expectResult: val.Address.String(),
		},
	}
	ownerURL := val.APIAddress + "/cosmos/nft/v1beta1/owner/%s/%s"
	for _, tc := range testCases {
		uri := fmt.Sprintf(ownerURL, tc.args.ClassID, tc.args.ID)
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(uri)
			if tc.expectErr {
				s.Require().Contains(string(resp), tc.errMsg)
			} else {
				s.Require().NoError(err)
				var result nft.QueryOwnerResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp, &result)
				s.Require().NoError(err)
				s.Require().EqualValues(tc.expectResult, result.Owner)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQuerySupplyGRPC() {
	val := s.network.Validators[0]

	testCases := []struct {
		name string
		args struct {
			ClassID string
		}
		expectErr    bool
		errMsg       string
		expectResult uint64
	}{
		{
			name: "class id is invalid",
			args: struct {
				ClassID string
			}{
				ClassID: "invalid_class_id",
			},
			expectErr:    true,
			errMsg:       "invalid class id",
			expectResult: 0,
		},
		{
			name: "class id does not exist",
			args: struct {
				ClassID string
			}{
				ClassID: "class-id",
			},
			expectErr:    false,
			expectResult: 0,
		},
		{
			name: "class id exist",
			args: struct {
				ClassID string
			}{
				ClassID: ExpNFT.ClassId,
			},
			expectErr:    false,
			expectResult: 1,
		},
	}
	supplyURL := val.APIAddress + "/cosmos/nft/v1beta1/supply/%s"
	for _, tc := range testCases {
		uri := fmt.Sprintf(supplyURL, tc.args.ClassID)
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

func (s *IntegrationTestSuite) TestQueryNFTsGRPC() {
	val := s.network.Validators[0]
	testCases := []struct {
		name string
		args struct {
			ClassID string
			Owner   string
		}
		expectErr    bool
		errorMsg     string
		expectResult []*nft.NFT
	}{
		{
			name: "classID and owner are both empty",
			args: struct {
				ClassID string
				Owner   string
			}{},
			errorMsg:     "must provide at least one of classID or owner",
			expectErr:    true,
			expectResult: []*nft.NFT{},
		},
		{
			name: "classID is invalid",
			args: struct {
				ClassID string
				Owner   string
			}{
				ClassID: "invalid_class_id",
			},
			expectErr:    true,
			expectResult: []*nft.NFT{},
		},
		{
			name: "classID does not exist",
			args: struct {
				ClassID string
				Owner   string
			}{
				ClassID: "class-id",
			},
			expectErr:    false,
			expectResult: []*nft.NFT{},
		},
		{
			name: "success query by classID",
			args: struct {
				ClassID string
				Owner   string
			}{
				ClassID: ExpNFT.ClassId,
			},
			expectErr:    false,
			expectResult: []*nft.NFT{&ExpNFT},
		},
		{
			name: "success query by owner",
			args: struct {
				ClassID string
				Owner   string
			}{
				Owner: val.Address.String(),
			},
			expectErr:    false,
			expectResult: []*nft.NFT{&ExpNFT},
		},
		{
			name: "success query by owner and classID",
			args: struct {
				ClassID string
				Owner   string
			}{
				ClassID: ExpNFT.ClassId,
				Owner:   val.Address.String(),
			},
			expectErr:    false,
			expectResult: []*nft.NFT{&ExpNFT},
		},
	}
	nftsOfClassURL := val.APIAddress + "/cosmos/nft/v1beta1/nfts?class_id=%s&owner=%s"
	for _, tc := range testCases {
		uri := fmt.Sprintf(nftsOfClassURL, tc.args.ClassID, tc.args.Owner)
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(uri)
			if tc.expectErr {
				s.Require().Contains(string(resp), tc.errorMsg)
			} else {
				s.Require().NoError(err)
				var result nft.QueryNFTsResponse
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
			ClassID string
			ID      string
		}
		expectErr bool
		errorMsg  string
	}{
		{
			name: "class id is invalid",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: "invalid_class_id",
				ID:      ExpNFT.Id,
			},
			expectErr: true,
			errorMsg:  "invalid class id",
		},
		{
			name: "class id does not exist",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: "class",
				ID:      ExpNFT.Id,
			},
			expectErr: true,
			errorMsg:  "not found nft",
		},
		{
			name: "nft id is invalid",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: ExpNFT.ClassId,
				ID:      "invalid_nft_id",
			},
			expectErr: true,
			errorMsg:  "invalid nft id",
		},
		{
			name: "nft id does not exist",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: ExpNFT.ClassId,
				ID:      "nft-id",
			},
			expectErr: true,
			errorMsg:  "not found nft",
		},
		{
			name: "exist nft",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: ExpNFT.ClassId,
				ID:      ExpNFT.Id,
			},
			expectErr: false,
		},
	}
	nftURL := val.APIAddress + "/cosmos/nft/v1beta1/nfts/%s/%s"
	for _, tc := range testCases {
		uri := fmt.Sprintf(nftURL, tc.args.ClassID, tc.args.ID)
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
			ClassID string
		}
		expectErr bool
		errorMsg  string
	}{
		{
			name: "class id does not exist",
			args: struct {
				ClassID string
			}{
				ClassID: "class-id",
			},
			expectErr: true,
			errorMsg:  "not found class",
		},
		{
			name: "class id exist",
			args: struct {
				ClassID string
			}{
				ClassID: ExpNFT.ClassId,
			},
			expectErr: false,
		},
	}
	classURL := val.APIAddress + "/cosmos/nft/v1beta1/classes/%s"
	for _, tc := range testCases {
		uri := fmt.Sprintf(classURL, tc.args.ClassID)
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
