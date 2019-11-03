package cli

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/stretchr/testify/assert"

	"github.com/tendermint/tendermint/rpc/client/mock"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MockClient struct {
	mock.Client
	heights []time.Time
}

func (client MockClient) Block(height *int64) (*ctypes.ResultBlock, error) {
	return &ctypes.ResultBlock{
		BlockMeta: &types.BlockMeta{
			Header: types.Header{
				Time: client.heights[*height],
			},
		},
	}, nil
}

func TestFindBlockHeightWithDate(t *testing.T) {
	startTime := time.Date(2019, 01, 01, 0, 0, 0, 0, time.Local)
	client := MockClient{
		heights: []time.Time{
			startTime,
			startTime.Add(20 * time.Second),
			startTime.Add(100 * time.Second),
			startTime.Add(200 * time.Second),
			startTime.Add(300 * time.Second)},
	}

	tests := []struct {
		startTime      time.Time
		expectedHeight int64
		error          bool
	}{
		{startTime, 0, true},
		{startTime.Add(20 * time.Second), int64(1), false},
		{startTime.Add(50 * time.Second), int64(2), false},
		{startTime.Add(150 * time.Second), int64(3), false},
		{startTime.Add(250 * time.Second), int64(4), false},
		{startTime.Add(350 * time.Second), 0, true},
	}

	for _, test := range tests {
		height, err := findBlockHeightFromDate(test.startTime, client, 4)
		if test.error {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expectedHeight, *height)
		}
	}
}

type MockValdaitorQuerier struct {
	valuePerBlock []map[string]sdk.DecCoins
}

func (m MockValdaitorQuerier) GetValidatorCommission(height int64, validatorAddress sdk.ValAddress) (sdk.DecCoins, error) {
	return m.valuePerBlock[height][validatorAddress.String()], nil
}

func (m MockValdaitorQuerier) GetValidatorRewards(height int64, validatorAddress sdk.ValAddress) (sdk.DecCoins, error) {
	return m.valuePerBlock[height][validatorAddress.String()], nil
}

func TestCalculateValidatorRewardsAndCommissionPerDay(t *testing.T) {
	startTime := time.Date(2019, 01, 01, 0, 0, 0, 0, time.Local)
	endTime := time.Date(2019, 01, 04, 0, 0, 0, 0, time.Local)
	times := []time.Time{
		startTime,
		startTime.Add(12 * time.Hour),
		startTime.Add(15 * time.Hour),
		startTime.Add(24 * time.Hour),
		startTime.Add(32 * time.Hour),
		startTime.Add(40 * time.Hour),
		startTime.Add(300 * time.Hour),
	}
	client := MockClient{
		heights: times,
	}
	var pub ed25519.PubKeyEd25519
	rand.Read(pub[:])
	validator1 := sdk.ValAddress(pub.Address())
	rand.Read(pub[:])
	validator2 := sdk.ValAddress(pub.Address())
	valuePerBlock := []map[string]sdk.DecCoins{
		{
			validator1.String(): {sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(0)}},
			validator2.String(): {sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(0)}},
		},
		{
			validator1.String(): {sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(1)}},
			validator2.String(): {sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(10)}},
		},
		{
			validator1.String(): {sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(5)}},
			validator2.String(): {sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(25)}},
		},
		{
			validator1.String(): {sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(55)}},
			validator2.String(): {sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(125)}},
		},
		{
			validator1.String(): {sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(65)}},
			validator2.String(): {sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(150)}},
		},
		{
			validator1.String(): {sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(85)}},
			validator2.String(): {sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(250)}},
		},
		{
			validator1.String(): {sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(115)}},
			validator2.String(): {sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(350)}},
		},
	}
	querier := MockValdaitorQuerier{
		valuePerBlock: valuePerBlock,
	}
	expected1 := [][]sdk.DecCoins{
		{
			{
				sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(5)},
			},
			{
				sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(5)},
			},
		},
		{
			{
				sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(80)},
			},
			{
				sdk.DecCoin{Denom: "atom", Amount: sdk.NewDec(80)},
			},
		},
	}
	aggByDay, days, err := calculateValidatorRewardsAndCommissionPerDay(client, validator1, querier, endTime, 1, 6)
	assert.NoError(t, err)
	assert.ElementsMatch(t, days, []time.Time{times[0], times[3]})
	assert.ElementsMatch(t, aggByDay, expected1)
}
