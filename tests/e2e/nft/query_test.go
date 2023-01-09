package nft

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/nft"
	"gotest.tools/v3/assert"
)

func TestQueryClass(t *testing.T) {
	t.Parallel()
	f, cleanup := initFixture(t)
	defer cleanup()

	val := f.network.Validators[0]
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
		t.Run(tc.name, func(t *testing.T) {
			resp, err := ExecQueryClass(val, tc.args.ClassID)
			if tc.expectErr {
				assert.ErrorContains(t, err, "")
			} else {
				assert.NilError(t, err)
				var result nft.QueryClassResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &result)
				assert.NilError(t, err)
				assert.DeepEqual(t, ExpClass, *result.Class)
			}
		})
	}
}

func TestQueryClasses(t *testing.T) {
	t.Parallel()
	f, cleanup := initFixture(t)
	defer cleanup()

	val := f.network.Validators[0]
	testCases := []struct {
		name string
	}{
		{
			name: "no params",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := ExecQueryClasses(val)
			assert.NilError(t, err)
			var result nft.QueryClassesResponse
			err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &result)
			assert.NilError(t, err)
			assert.Assert(t, len(result.Classes) == 1)
			assert.DeepEqual(t, ExpClass, *result.Classes[0])
		})
	}
}

func TestQueryNFT(t *testing.T) {
	t.Parallel()
	f, cleanup := initFixture(t)
	defer cleanup()

	val := f.network.Validators[0]
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
		t.Run(tc.name, func(t *testing.T) {
			resp, err := ExecQueryNFT(val, tc.args.ClassID, tc.args.ID)
			if tc.expectErr {
				assert.NilError(t, err)
			} else {
				assert.NilError(t, err)
				var result nft.QueryNFTResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &result)
				assert.NilError(t, err)
				assert.DeepEqual(t, ExpNFT, *result.Nft)
			}
		})
	}
}

func TestQueryNFTs(t *testing.T) {
	t.Parallel()
	f, cleanup := initFixture(t)
	defer cleanup()

	val := f.network.Validators[0]
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
				Owner:   f.owner.String(),
			},
			expectErr:    false,
			expectResult: []*nft.NFT{},
		},
		{
			name: "class id and owner both does not exist",
			args: struct {
				ClassID string
				Owner   string
			}{},
			expectErr:    true,
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
		t.Run(tc.name, func(t *testing.T) {
			resp, err := ExecQueryNFTs(val, tc.args.ClassID, tc.args.Owner)
			if tc.expectErr {
				assert.NilError(t, err)
			} else {
				assert.NilError(t, err)
				var result nft.QueryNFTsResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &result)
				assert.NilError(t, err)
				assert.DeepEqual(t, tc.expectResult, result.Nfts)
			}
		})
	}
}

func TestQueryOwner(t *testing.T) {
	t.Parallel()
	f, cleanup := initFixture(t)
	defer cleanup()

	val := f.network.Validators[0]
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
		t.Run(tc.name, func(t *testing.T) {
			resp, err := ExecQueryOwner(val, tc.args.ClassID, tc.args.ID)
			if tc.expectErr {
				assert.ErrorContains(t, err, "")
				assert.Assert(t, strings.Contains(string(resp.Bytes()), tc.errorMsg))
			} else {
				assert.NilError(t, err)
				var result nft.QueryOwnerResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &result)
				assert.NilError(t, err)
				assert.DeepEqual(t, tc.expectResult, result.Owner)
			}
		})
	}
}

func TestQueryBalance(t *testing.T) {
	t.Parallel()
	f, cleanup := initFixture(t)
	defer cleanup()

	val := f.network.Validators[0]
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
				Owner:   f.owner.String(),
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
		t.Run(tc.name, func(t *testing.T) {
			resp, err := ExecQueryBalance(val, tc.args.ClassID, tc.args.Owner)
			if tc.expectErr {
				assert.ErrorContains(t, err, "")
			} else {
				assert.NilError(t, err)
				var result nft.QueryBalanceResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &result)
				assert.NilError(t, err)
				assert.DeepEqual(t, tc.expectResult, result.Amount)
			}
		})
	}
}

func TestQuerySupply(t *testing.T) {
	t.Parallel()
	f, cleanup := initFixture(t)
	defer cleanup()

	val := f.network.Validators[0]
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
			name: "class id is empty",
			args: struct {
				ClassID string
			}{
				ClassID: "",
			},
			expectErr:    true,
			errorMsg:     nft.ErrEmptyClassID.Error(),
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
		t.Run(tc.name, func(t *testing.T) {
			resp, err := ExecQuerySupply(val, tc.args.ClassID)
			if tc.expectErr {
				assert.ErrorContains(t, err, "")
			} else {
				assert.NilError(t, err)
				var result nft.QuerySupplyResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &result)
				assert.NilError(t, err)
				assert.DeepEqual(t, tc.expectResult, result.Amount)
			}
		})
	}
}
