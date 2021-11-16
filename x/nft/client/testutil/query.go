package testutil

import (
	"github.com/cosmos/cosmos-sdk/x/nft"
)

func (s *IntegrationTestSuite) TestQueryClass() {
	val := s.network.Validators[0]
	testCases := []struct {
		name string
		args struct {
			ClassID string
		}
		expectErr bool
	}{
		{
			name: "class id does not exist",
			args: struct {
				ClassID string
			}{
				ClassID: "class",
			},
			expectErr: true,
		},
		{
			name: "class id exist",
			args: struct {
				ClassID string
			}{
				ClassID: testClassID,
			},
			expectErr: false,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := ExecQueryClass(val, tc.args.ClassID)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var result nft.QueryClassResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &result)
				s.Require().NoError(err)
				s.Require().EqualValues(ExpClass, *result.Class)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryClasses() {
	val := s.network.Validators[0]
	testCases := []struct {
		name      string
		expectErr bool
	}{
		{
			name:      "no params",
			expectErr: false,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := ExecQueryClasses(val)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var result nft.QueryClassesResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &result)
				s.Require().NoError(err)
				s.Require().Len(result.Classes, 1)
				s.Require().EqualValues(ExpClass, *result.Classes[0])
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryNFT() {
	val := s.network.Validators[0]
	testCases := []struct {
		name string
		args struct {
			ClassID string
			ID      string
		}
		expectErr bool
	}{
		{
			name: "class id does not exist",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: "class",
				ID:      testID,
			},
			expectErr: true,
		},
		{
			name: "nft id does not exist",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: testClassID,
				ID:      "id",
			},
			expectErr: true,
		},
		{
			name: "exist nft",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: testClassID,
				ID:      testID,
			},
			expectErr: false,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := ExecQueryNFT(val, tc.args.ClassID, tc.args.ID)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var result nft.QueryNFTResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &result)
				s.Require().NoError(err)
				s.Require().EqualValues(ExpNFT, *result.Nft)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryNFTs() {
	val := s.network.Validators[0]
	testCases := []struct {
		name string
		args struct {
			ClassID string
		}
		expectErr    bool
		expectResult []*nft.NFT
	}{
		{
			name: "class id does not exist",
			args: struct {
				ClassID string
			}{
				ClassID: "class",
			},
			expectErr:    false,
			expectResult: []*nft.NFT{},
		},
		{
			name: "class id exist",
			args: struct {
				ClassID string
			}{
				ClassID: testClassID,
			},
			expectErr:    false,
			expectResult: []*nft.NFT{&ExpNFT},
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := ExecQueryNFTs(val, tc.args.ClassID)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var result nft.QueryNFTsOfClassResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &result)
				s.Require().NoError(err)
				s.Require().EqualValues(tc.expectResult, result.Nfts)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryNFTsByOwner() {
	val := s.network.Validators[0]
	testCases := []struct {
		name string
		args struct {
			ClassID string
			Owner   string
		}
		expectErr    bool
		expectResult []*nft.NFT
	}{
		{
			name: "class id does not exist",
			args: struct {
				ClassID string
				Owner   string
			}{
				ClassID: "class",
				Owner:   val.Address.String(),
			},
			expectErr:    false,
			expectResult: []*nft.NFT{},
		},
		{
			name: "owner does not exist",
			args: struct {
				ClassID string
				Owner   string
			}{
				ClassID: testClassID,
				Owner:   s.owner.String(),
			},
			expectErr:    false,
			expectResult: []*nft.NFT{},
		},
		{
			name: "nft exist",
			args: struct {
				ClassID string
				Owner   string
			}{
				ClassID: testClassID,
				Owner:   val.Address.String(),
			},
			expectErr:    false,
			expectResult: []*nft.NFT{&ExpNFT},
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := ExecQueryNFTsByOwner(val, tc.args.ClassID, tc.args.Owner)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var result nft.QueryNFTsOfClassResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &result)
				s.Require().NoError(err)
				s.Require().EqualValues(tc.expectResult, result.Nfts)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryOwner() {
	val := s.network.Validators[0]
	testCases := []struct {
		name string
		args struct {
			ClassID string
			ID      string
		}
		expectErr    bool
		errorMsg     string
		expectResult string
	}{
		{
			name: "class id is invalid",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: "invalid_class_id",
				ID:      testID,
			},
			expectErr:    true,
			errorMsg:     "invalid class id",
			expectResult: "",
		},
		{
			name: "class id does not exist",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: "class",
				ID:      testID,
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
				ClassID: testClassID,
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
				ClassID: testClassID,
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
				ClassID: testClassID,
				ID:      testID,
			},
			expectErr:    false,
			expectResult: val.Address.String(),
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := ExecQueryOwner(val, tc.args.ClassID, tc.args.ID)
			if tc.expectErr {
				s.Require().Contains(string(resp.Bytes()), tc.errorMsg)
			} else {
				s.Require().NoError(err)
				var result nft.QueryOwnerResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &result)
				s.Require().NoError(err)
				s.Require().EqualValues(tc.expectResult, result.Owner)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryBalance() {
	val := s.network.Validators[0]
	testCases := []struct {
		name string
		args struct {
			ClassID string
			Owner   string
		}
		expectErr    bool
		errorMsg     string
		expectResult uint64
	}{
		{
			name: "class id is invalid",
			args: struct {
				ClassID string
				Owner   string
			}{
				ClassID: "invalid_class_id",
				Owner:   val.Address.String(),
			},
			expectErr:    true,
			errorMsg:     "invalid class id",
			expectResult: 0,
		},
		{
			name: "class id does not exist",
			args: struct {
				ClassID string
				Owner   string
			}{
				ClassID: "class",
				Owner:   val.Address.String(),
			},
			expectErr:    false,
			expectResult: 0,
		},
		{
			name: "owner does not exist",
			args: struct {
				ClassID string
				Owner   string
			}{
				ClassID: testClassID,
				Owner:   s.owner.String(),
			},
			expectErr:    false,
			expectResult: 0,
		},
		{
			name: "nft exist",
			args: struct {
				ClassID string
				Owner   string
			}{
				ClassID: testClassID,
				Owner:   val.Address.String(),
			},
			expectErr:    false,
			expectResult: 1,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := ExecQueryBalance(val, tc.args.ClassID, tc.args.Owner)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var result nft.QueryBalanceResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &result)
				s.Require().NoError(err)
				s.Require().EqualValues(tc.expectResult, result.Amount)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQuerySupply() {
	val := s.network.Validators[0]
	testCases := []struct {
		name string
		args struct {
			ClassID string
		}
		expectErr    bool
		errorMsg     string
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
			errorMsg:     "invalid class id",
			expectResult: 0,
		},
		{
			name: "class id does not exist",
			args: struct {
				ClassID string
			}{
				ClassID: "class",
			},
			expectErr:    false,
			expectResult: 0,
		},
		{
			name: "class id exist",
			args: struct {
				ClassID string
			}{
				ClassID: testClassID,
			},
			expectErr:    false,
			expectResult: 1,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := ExecQuerySupply(val, tc.args.ClassID)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var result nft.QuerySupplyResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &result)
				s.Require().NoError(err)
				s.Require().EqualValues(tc.expectResult, result.Amount)
			}
		})
	}
}
