package simapp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// NewSimAppUNSAFE is used for debugging purposes only.
//
// NOTE: to not use this function with non-test code
func NewSimAppUNSAFE(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, baseAppOptions ...func(*bam.BaseApp),
) (gapp *SimApp, keyMain, keyStaking *sdk.KVStoreKey, stakingKeeper staking.Keeper) {

	gapp = NewSimApp(logger, db, traceStore, loadLatest, invCheckPeriod, baseAppOptions...)
	return gapp, gapp.keyMain, gapp.keyStaking, gapp.stakingKeeper
}

// getSimulationLog unmarshals the KVPair's Value to the corresponding type based on the
// each's module store key and the prefix bytes of the KVPair's key.
func getSimulationLog(storeName string, cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) (log string) {
	log = fmt.Sprintf("store A %X => %X\n"+
		"store B %X => %X", kvA.Key, kvA.Value, kvB.Key, kvB.Value)

	if len(kvA.Value) == 0 && len(kvB.Value) == 0 {
		return
	}

	switch storeName {
	case authtypes.StoreKey:
		return decodeAccountStore(cdcA, cdcB, kvA.Value, kvB.Value)
	case mint.StoreKey:
		return decodeMintStore(cdcA, cdcB, kvA.Value, kvB.Value)
	case staking.StoreKey:
		return decodeStakingStore(cdcA, cdcB, kvA, kvB)
	case gov.StoreKey:
		return decodeGovStore(cdcA, cdcB, kvA, kvB)
	case distribution.StoreKey:
		return decodeDistributionStore(cdcA, cdcB, kvA, kvB)
	default:
		return
	}
}

func decodeAccountStore(cdcA, cdcB *codec.Codec, valueA, valueB []byte) string {
	var accA, accB authtypes.Account
	cdcA.MustUnmarshalBinaryBare(valueA, &accA)
	cdcB.MustUnmarshalBinaryBare(valueB, &accB)
	return fmt.Sprintf("%v\n%v", accA, accB)
}

func decodeMintStore(cdcA, cdcB *codec.Codec, valueA, valueB []byte) string {
	var minterA, minterB mint.Minter
	cdcA.MustUnmarshalBinaryLengthPrefixed(valueA, &minterA)
	cdcB.MustUnmarshalBinaryLengthPrefixed(valueB, &minterB)
	return fmt.Sprintf("%v\n%v", minterA, minterB)
}

func decodeDistributionStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) (log string) {
	switch {
	case bytes.Equal(kvA.Key[:1], distribution.FeePoolKey):
		var feePoolA, feePoolB distribution.FeePool
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &feePoolA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &feePoolB)
		log = fmt.Sprintf("%v\n%v", feePoolA, feePoolB)

	case bytes.Equal(kvA.Key[:1], distribution.ProposerKey):
		log = fmt.Sprintf("%v\n%v", sdk.ConsAddress(kvA.Value), sdk.ConsAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorOutstandingRewardsPrefix):
		var rewardsA, rewardsB distribution.ValidatorOutstandingRewards
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		log = fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], distribution.DelegatorWithdrawAddrPrefix):
		log = fmt.Sprintf("%v\n%v", sdk.AccAddress(kvA.Value), sdk.AccAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], distribution.DelegatorStartingInfoPrefix):
		var infoA, infoB distribution.DelegatorStartingInfo
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &infoA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &infoB)
		log = fmt.Sprintf("%v\n%v", infoA, infoB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorHistoricalRewardsPrefix):
		var rewardsA, rewardsB distribution.ValidatorHistoricalRewards
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		log = fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorCurrentRewardsPrefix):
		var rewardsA, rewardsB distribution.ValidatorCurrentRewards
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		log = fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorAccumulatedCommissionPrefix):
		var commissionA, commissionB distribution.ValidatorAccumulatedCommission
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &commissionA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &commissionB)
		log = fmt.Sprintf("%v\n%v", commissionA, commissionB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorSlashEventPrefix):
		var eventA, eventB distribution.ValidatorSlashEvent
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &eventA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &eventB)
		log = fmt.Sprintf("%v\n%v", eventA, eventB)

	default:
		panic(fmt.Sprintf("invalid key prefix %X", kvA.Key[:1]))

	}
	return log
}

func decodeStakingStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) (log string) {
	switch {
	case bytes.Equal(kvA.Key[:1], staking.PoolKey):
		var poolA, poolB staking.Pool
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &poolA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &poolB)
		log = fmt.Sprintf("%v\n%v", poolA, poolB)

	case bytes.Equal(kvA.Key[:1], staking.LastTotalPowerKey):
		var powerA, powerB sdk.Int
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &powerA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &powerB)
		log = fmt.Sprintf("%v\n%v", powerA, powerB)

	case bytes.Equal(kvA.Key[:1], staking.ValidatorsKey):
		var validatorA, validatorB staking.Validator
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &validatorA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &validatorB)
		log = fmt.Sprintf("%v\n%v", validatorA, validatorB)

	case bytes.Equal(kvA.Key[:1], staking.LastValidatorPowerKey),
		bytes.Equal(kvA.Key[:1], staking.ValidatorsByConsAddrKey),
		bytes.Equal(kvA.Key[:1], staking.ValidatorsByPowerIndexKey):
		log = fmt.Sprintf("%v\n%v", sdk.ValAddress(kvA.Value), sdk.ValAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], staking.DelegationKey):
		var delegationA, delegationB staking.Delegation
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &delegationA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &delegationB)
		log = fmt.Sprintf("%v\n%v", delegationA, delegationB)

	case bytes.Equal(kvA.Key[:1], staking.UnbondingDelegationKey),
		bytes.Equal(kvA.Key[:1], staking.UnbondingDelegationByValIndexKey):
		var ubdA, ubdB staking.UnbondingDelegation
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &ubdA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &ubdB)
		log = fmt.Sprintf("%v\n%v", ubdA, ubdB)

	case bytes.Equal(kvA.Key[:1], staking.RedelegationKey),
		bytes.Equal(kvA.Key[:1], staking.RedelegationByValSrcIndexKey):
		var redA, redB staking.Redelegation
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &redA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &redB)
		log = fmt.Sprintf("%v\n%v", redA, redB)

	default:
		panic(fmt.Sprintf("invalid key prefix %X", kvA.Key[:1]))

	}
	return log
}

func decodeGovStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) (log string) {
	switch {
	case bytes.Equal(kvA.Key[:1], gov.ProposalsKeyPrefix):
		var proposalA, proposalB gov.Proposal
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &proposalA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &proposalB)
		log = fmt.Sprintf("%v\n%v", proposalA, proposalB)

	case bytes.Equal(kvA.Key[:1], gov.ActiveProposalQueuePrefix),
		bytes.Equal(kvA.Key[:1], gov.InactiveProposalQueuePrefix),
		bytes.Equal(kvA.Key[:1], gov.ProposalIDKey):
		proposalIDA := binary.LittleEndian.Uint64(kvA.Value)
		proposalIDB := binary.LittleEndian.Uint64(kvB.Value)
		log = fmt.Sprintf("proposalIDA: %d\nProposalIDB: %d", proposalIDA, proposalIDB)

	case bytes.Equal(kvA.Key[:1], gov.DepositsKeyPrefix):
		var depositA, depositB gov.Deposit
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &depositA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &depositB)
		log = fmt.Sprintf("%v\n%v", depositA, depositB)

	case bytes.Equal(kvA.Key[:1], gov.VotesKeyPrefix):
		var voteA, voteB gov.Vote
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &voteA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &voteB)
		log = fmt.Sprintf("%v\n%v", voteA, voteB)

	default:
		panic(fmt.Sprintf("invalid key prefix %X", kvA.Key[:1]))

	}
	return log
}
