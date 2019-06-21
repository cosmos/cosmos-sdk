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

func retrieveSimLog(storeName string, appA, appB SimApp, kvA, kvB cmn.KVPair) (log string) {
	log = fmt.Sprintf("store A %X => %X\n"+
		"store B %X => %X", kvA.Key, kvA.Value, kvB.Key, kvB.Value)

	if len(kvA.Value) == 0 && len(kvB.Value) == 0 {
		return
	}

	switch storeName {
	case authtypes.StoreKey:
		return decodeAccountStore(appA.cdc, appB.cdc, kvA.Value, kvB.Value)
	case mint.StoreKey:
		return decodeMintStore(appA.cdc, appB.cdc, kvA.Value, kvB.Value)
	case staking.StoreKey:
		return decodeStakingStore(appA.cdc, appB.cdc, kvA, kvB)
	case gov.StoreKey:
		return decodeGovStore(appA.cdc, appB.cdc, kvA, kvB)
	case distribution.StoreKey:
		return decodeDistributionStore(appA.cdc, appB.cdc, kvA, kvB)
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
	cdcA.MustUnmarshalBinaryBare(valueA, &minterA)
	cdcB.MustUnmarshalBinaryBare(valueB, &minterB)
	return fmt.Sprintf("%v\n%v", minterA, minterA)
}

func decodeDistributionStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) (log string) {
	switch {
	case bytes.Equal(kvA.Key[:1], distribution.FeePoolKey):
		var feePoolA, feePoolB distribution.FeePool
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &feePoolA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &feePoolB)
		log = fmt.Sprintf("%v\n%v", feePoolA, feePoolB)

	case bytes.Equal(kvA.Key[:1], distribution.ProposerKey):
		var rewardsA, rewardsB distribution.ValidatorOutstandingRewards
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &rewardsA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &rewardsB)
		log = fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], distribution.DelegatorWithdrawAddrPrefix):
		var addrA, addrB gov.Deposit
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &addrA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &addrB)
		log = fmt.Sprintf("%v\n%v", addrA, addrB)

	case bytes.Equal(kvA.Key[:1], distribution.DelegatorStartingInfoPrefix):
		var infoA, infoB distribution.DelegatorStartingInfo
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &infoA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &infoB)
		log = fmt.Sprintf("%v\n%v", infoA, infoB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorHistoricalRewardsPrefix):
		var infoA, infoB distribution.DelegatorStartingInfo
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &infoA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &infoB)
		log = fmt.Sprintf("%v\n%v", infoA, infoB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorCurrentRewardsPrefix):
		var rewardsA, rewardsB distribution.ValidatorCurrentRewards
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &rewardsA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &rewardsB)
		log = fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorAccumulatedCommissionPrefix):
		var commissionA, commissionB distribution.ValidatorAccumulatedCommission
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &commissionA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &commissionB)
		log = fmt.Sprintf("%v\n%v", commissionA, commissionB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorSlashEventPrefix):
		var eventA, eventB distribution.ValidatorSlashEvent
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &eventA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &eventB)
		log = fmt.Sprintf("%v\n%v", eventA, eventB)

	default:
		panic(fmt.Sprintf("invalid key prefix %X", kvA.Key[:1]))

	}
	return log
}

func decodeStakingStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) (log string) {
	switch {
	case bytes.Equal(kvA.Key[:1], staking.PoolKey):
		var feePoolA, feePoolB staking.Pool
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &feePoolA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &feePoolB)
		log = fmt.Sprintf("%v\n%v", feePoolA, feePoolB)

	case bytes.Equal(kvA.Key[:1], staking.LastTotalPowerKey):
		var powerA, powerB sdk.Int
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &powerA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &powerB)
		log = fmt.Sprintf("%v\n%v", powerA, powerB)

	case bytes.Equal(kvA.Key[:1], staking.ValidatorsKey):
		var validatorA, validatorB staking.Validator
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &validatorA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &validatorB)
		log = fmt.Sprintf("%v\n%v", validatorA, validatorB)

	case bytes.Equal(kvA.Key[:1], staking.LastValidatorPowerKey),
		bytes.Equal(kvA.Key[:1], staking.ValidatorsByConsAddrKey),
		bytes.Equal(kvA.Key[:1], staking.ValidatorsByPowerIndexKey):
		var valAddrA, valAddrB sdk.ValAddress
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &valAddrA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &valAddrB)
		log = fmt.Sprintf("%v\n%v", valAddrA, valAddrB)

	case bytes.Equal(kvA.Key[:1], staking.DelegationKey):
		var delegationA, delegationB staking.Delegation
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &delegationA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &delegationB)
		log = fmt.Sprintf("%v\n%v", delegationA, delegationB)

	case bytes.Equal(kvA.Key[:1], staking.UnbondingDelegationKey),
		bytes.Equal(kvA.Key[:1], staking.UnbondingDelegationByValIndexKey):
		var ubdA, ubdB staking.UnbondingDelegation
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &ubdA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &ubdB)
		log = fmt.Sprintf("%v\n%v", ubdA, ubdB)

	case bytes.Equal(kvA.Key[:1], staking.RedelegationKey),
		bytes.Equal(kvA.Key[:1], staking.RedelegationByValSrcIndexKey):
		var redA, redB staking.Redelegation
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &redA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &redB)
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
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &proposalA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &proposalB)
		log = fmt.Sprintf("%v\n%v", proposalA, proposalB)

	case bytes.Equal(kvA.Key[:1], gov.ActiveProposalQueuePrefix),
		bytes.Equal(kvA.Key[:1], gov.InactiveProposalQueuePrefix),
		bytes.Equal(kvA.Key[:1], gov.ProposalIDKey):
		proposalIDA := binary.LittleEndian.Uint64(kvA.Value)
		proposalIDB := binary.LittleEndian.Uint64(kvB.Value)
		log = fmt.Sprintf("proposalIDA: %d\nProposalIDB: %d", proposalIDA, proposalIDB)

	case bytes.Equal(kvA.Key[:1], gov.DepositsKeyPrefix):
		var depositA, depositB gov.Deposit
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &depositA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &depositB)
		log = fmt.Sprintf("%v\n%v", depositA, depositB)

	case bytes.Equal(kvA.Key[:1], gov.VotesKeyPrefix):
		var voteA, voteB gov.Vote
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &voteA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &voteB)
		log = fmt.Sprintf("%v\n%v", voteA, voteB)

	default:
		panic(fmt.Sprintf("invalid key prefix %X", kvA.Key[:1]))

	}
	return log
}
