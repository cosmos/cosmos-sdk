package simapp

import (
	"encoding/binary"
	"fmt"
	"time"

	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	delPk1    = ed25519.GenPrivKey().PubKey()
	delAddr1  = sdk.AccAddress(delPk1.Address())
	valAddr1  = sdk.ValAddress(delPk1.Address())
	consAddr1 = sdk.ConsAddress(delPk1.Address().Bytes())
)

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	auth.RegisterCodec(cdc)
	distr.RegisterCodec(cdc)
	gov.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	supply.RegisterCodec(cdc)
	return
}

func TestGetSimulationLog(t *testing.T) {
	cdc := makeTestCodec()

	tests := []struct {
		store  string
		kvPair cmn.KVPair
	}{
		{auth.StoreKey, cmn.KVPair{Key: auth.AddressStoreKey(delAddr1), Value: cdc.MustMarshalBinaryBare(auth.BaseAccount{})}},
		{mint.StoreKey, cmn.KVPair{Key: mint.MinterKey, Value: cdc.MustMarshalBinaryLengthPrefixed(mint.Minter{})}},
		{staking.StoreKey, cmn.KVPair{Key: staking.LastValidatorPowerKey, Value: valAddr1.Bytes()}},
		{gov.StoreKey, cmn.KVPair{Key: gov.VoteKey(1, delAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(gov.Vote{})}},
		{distribution.StoreKey, cmn.KVPair{Key: distr.ProposerKey, Value: consAddr1.Bytes()}},
		{slashing.StoreKey, cmn.KVPair{Key: slashing.GetValidatorMissedBlockBitArrayKey(consAddr1, 6), Value: cdc.MustMarshalBinaryLengthPrefixed(true)}},
		{supply.StoreKey, cmn.KVPair{Key: supply.SupplyKey, Value: cdc.MustMarshalBinaryLengthPrefixed(supply.NewSupply(sdk.Coins{}))}},
		{"Empty", cmn.KVPair{}},
		{"OtherStore", cmn.KVPair{Key: []byte("key"), Value: []byte("value")}},
	}

	for _, tt := range tests {
		t.Run(tt.store, func(t *testing.T) {
			require.NotPanics(t, func() { GetSimulationLog(tt.store, cdc, cdc, tt.kvPair, tt.kvPair) }, tt.store)
		})
	}
}

func TestDecodeAccountStore(t *testing.T) {
	cdc := makeTestCodec()
	acc := auth.NewBaseAccountWithAddress(delAddr1)
	globalAccNumber := uint64(10)

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: auth.AddressStoreKey(delAddr1), Value: cdc.MustMarshalBinaryBare(acc)},
		cmn.KVPair{Key: auth.GlobalAccountNumberKey, Value: cdc.MustMarshalBinaryLengthPrefixed(globalAccNumber)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
	}
	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Minter", fmt.Sprintf("%v\n%v", acc, acc)},
		{"GlobalAccNumber", fmt.Sprintf("GlobalAccNumberA: %d\nGlobalAccNumberB: %d", globalAccNumber, globalAccNumber)},
		{"other", ""},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { DecodeAccountStore(cdc, cdc, kvPairs[i], kvPairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, DecodeAccountStore(cdc, cdc, kvPairs[i], kvPairs[i]), tt.name)
			}
		})
	}
}

func TestDecodeMintStore(t *testing.T) {
	cdc := makeTestCodec()
	minter := mint.NewMinter(sdk.OneDec(), sdk.NewDec(15))

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: mint.MinterKey, Value: cdc.MustMarshalBinaryLengthPrefixed(minter)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
	}
	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Minter", fmt.Sprintf("%v\n%v", minter, minter)},
		{"other", ""},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { DecodeMintStore(cdc, cdc, kvPairs[i], kvPairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, DecodeMintStore(cdc, cdc, kvPairs[i], kvPairs[i]), tt.name)
			}
		})
	}
}

func TestDecodeDistributionStore(t *testing.T) {
	cdc := makeTestCodec()

	decCoins := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.OneDec())}
	feePool := distr.InitialFeePool()
	feePool.CommunityPool = decCoins
	info := distr.NewDelegatorStartingInfo(2, sdk.OneDec(), 200)
	outstanding := distr.ValidatorOutstandingRewards{decCoins[0]}
	commission := distr.ValidatorAccumulatedCommission{decCoins[0]}
	historicalRewards := distr.NewValidatorHistoricalRewards(decCoins, 100)
	currentRewards := distr.NewValidatorCurrentRewards(decCoins, 5)
	slashEvent := distr.NewValidatorSlashEvent(10, sdk.OneDec())

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: distr.FeePoolKey, Value: cdc.MustMarshalBinaryLengthPrefixed(feePool)},
		cmn.KVPair{Key: distr.ProposerKey, Value: consAddr1.Bytes()},
		cmn.KVPair{Key: distr.GetValidatorOutstandingRewardsKey(valAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(outstanding)},
		cmn.KVPair{Key: distr.GetDelegatorWithdrawAddrKey(delAddr1), Value: delAddr1.Bytes()},
		cmn.KVPair{Key: distr.GetDelegatorStartingInfoKey(valAddr1, delAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(info)},
		cmn.KVPair{Key: distr.GetValidatorHistoricalRewardsKey(valAddr1, 100), Value: cdc.MustMarshalBinaryLengthPrefixed(historicalRewards)},
		cmn.KVPair{Key: distr.GetValidatorCurrentRewardsKey(valAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(currentRewards)},
		cmn.KVPair{Key: distr.GetValidatorAccumulatedCommissionKey(valAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(commission)},
		cmn.KVPair{Key: distr.GetValidatorSlashEventKeyPrefix(valAddr1, 13), Value: cdc.MustMarshalBinaryLengthPrefixed(slashEvent)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"FeePool", fmt.Sprintf("%v\n%v", feePool, feePool)},
		{"Proposer", fmt.Sprintf("%v\n%v", consAddr1, consAddr1)},
		{"ValidatorOutstandingRewards", fmt.Sprintf("%v\n%v", outstanding, outstanding)},
		{"DelegatorWithdrawAddr", fmt.Sprintf("%v\n%v", delAddr1, delAddr1)},
		{"DelegatorStartingInfo", fmt.Sprintf("%v\n%v", info, info)},
		{"ValidatorHistoricalRewards", fmt.Sprintf("%v\n%v", historicalRewards, historicalRewards)},
		{"ValidatorCurrentRewards", fmt.Sprintf("%v\n%v", currentRewards, currentRewards)},
		{"ValidatorAccumulatedCommission", fmt.Sprintf("%v\n%v", commission, commission)},
		{"ValidatorSlashEvent", fmt.Sprintf("%v\n%v", slashEvent, slashEvent)},
		{"other", ""},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { DecodeDistributionStore(cdc, cdc, kvPairs[i], kvPairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, DecodeDistributionStore(cdc, cdc, kvPairs[i], kvPairs[i]), tt.name)
			}
		})
	}
}

func TestDecodeStakingStore(t *testing.T) {
	cdc := makeTestCodec()

	bondTime := time.Now().UTC()

	val := staking.NewValidator(valAddr1, delPk1, staking.NewDescription("test", "test", "test", "test"))
	del := staking.NewDelegation(delAddr1, valAddr1, sdk.OneDec())
	ubd := staking.NewUnbondingDelegation(delAddr1, valAddr1, 15, bondTime, sdk.OneInt())
	red := staking.NewRedelegation(delAddr1, valAddr1, valAddr1, 12, bondTime, sdk.OneInt(), sdk.OneDec())

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: staking.LastTotalPowerKey, Value: cdc.MustMarshalBinaryLengthPrefixed(sdk.OneInt())},
		cmn.KVPair{Key: staking.GetValidatorKey(valAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(val)},
		cmn.KVPair{Key: staking.LastValidatorPowerKey, Value: valAddr1.Bytes()},
		cmn.KVPair{Key: staking.GetDelegationKey(delAddr1, valAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(del)},
		cmn.KVPair{Key: staking.GetUBDKey(delAddr1, valAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(ubd)},
		cmn.KVPair{Key: staking.GetREDKey(delAddr1, valAddr1, valAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(red)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"LastTotalPower", fmt.Sprintf("%v\n%v", sdk.OneInt(), sdk.OneInt())},
		{"Validator", fmt.Sprintf("%v\n%v", val, val)},
		{"LastValidatorPower/ValidatorsByConsAddr/ValidatorsByPowerIndex", fmt.Sprintf("%v\n%v", valAddr1, valAddr1)},
		{"Delegation", fmt.Sprintf("%v\n%v", del, del)},
		{"UnbondingDelegation", fmt.Sprintf("%v\n%v", ubd, ubd)},
		{"Redelegation", fmt.Sprintf("%v\n%v", red, red)},
		{"other", ""},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { DecodeStakingStore(cdc, cdc, kvPairs[i], kvPairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, DecodeStakingStore(cdc, cdc, kvPairs[i], kvPairs[i]), tt.name)
			}
		})
	}
}

func TestDecodeSlashingStore(t *testing.T) {
	cdc := makeTestCodec()

	info := slashing.NewValidatorSigningInfo(consAddr1, 0, 1, time.Now().UTC(), false, 0)
	bechPK := sdk.MustBech32ifyAccPub(delPk1)
	missed := true

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: slashing.GetValidatorSigningInfoKey(consAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(info)},
		cmn.KVPair{Key: slashing.GetValidatorMissedBlockBitArrayKey(consAddr1, 6), Value: cdc.MustMarshalBinaryLengthPrefixed(missed)},
		cmn.KVPair{Key: slashing.GetAddrPubkeyRelationKey(delAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(delPk1)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"ValidatorSigningInfo", fmt.Sprintf("%v\n%v", info, info)},
		{"ValidatorMissedBlockBitArray", fmt.Sprintf("missedA: %v\nmissedB: %v", missed, missed)},
		{"AddrPubkeyRelation", fmt.Sprintf("PubKeyA: %s\nPubKeyB: %s", bechPK, bechPK)},
		{"other", ""},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { DecodeSlashingStore(cdc, cdc, kvPairs[i], kvPairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, DecodeSlashingStore(cdc, cdc, kvPairs[i], kvPairs[i]), tt.name)
			}
		})
	}
}

func TestDecodeGovStore(t *testing.T) {
	cdc := makeTestCodec()

	endTime := time.Now().UTC()

	content := gov.ContentFromProposalType("test", "test", gov.ProposalTypeText)
	proposal := gov.NewProposal(content, 1, endTime, endTime.Add(24*time.Hour))
	proposalIDBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(proposalIDBz, 1)
	deposit := gov.NewDeposit(1, delAddr1, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())))
	vote := gov.NewVote(1, delAddr1, gov.OptionYes)

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: gov.ProposalKey(1), Value: cdc.MustMarshalBinaryLengthPrefixed(proposal)},
		cmn.KVPair{Key: gov.InactiveProposalQueueKey(1, endTime), Value: proposalIDBz},
		cmn.KVPair{Key: gov.DepositKey(1, delAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(deposit)},
		cmn.KVPair{Key: gov.VoteKey(1, delAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(vote)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"proposals", fmt.Sprintf("%v\n%v", proposal, proposal)},
		{"proposal IDs", "proposalIDA: 1\nProposalIDB: 1"},
		{"deposits", fmt.Sprintf("%v\n%v", deposit, deposit)},
		{"votes", fmt.Sprintf("%v\n%v", vote, vote)},
		{"other", ""},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { DecodeGovStore(cdc, cdc, kvPairs[i], kvPairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, DecodeGovStore(cdc, cdc, kvPairs[i], kvPairs[i]), tt.name)
			}
		})
	}
}

func TestDecodeSupplyStore(t *testing.T) {
	cdc := makeTestCodec()

	totalSupply := supply.NewSupply(sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)))

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: supply.SupplyKey, Value: cdc.MustMarshalBinaryLengthPrefixed(totalSupply)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Supply", fmt.Sprintf("%v\n%v", totalSupply, totalSupply)},
		{"other", ""},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { DecodeSupplyStore(cdc, cdc, kvPairs[i], kvPairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, DecodeSupplyStore(cdc, cdc, kvPairs[i], kvPairs[i]), tt.name)
			}
		})
	}
}
