//nolint
package simapp

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

//---------------------------------------------------------------------
// Flags

// List of available flags for the simulator
var (
	genesisFile        string
	paramsFile         string
	exportParamsPath   string
	exportParamsHeight int
	exportStatePath    string
	exportStatsPath    string
	seed               int64
	initialBlockHeight int
	numBlocks          int
	blockSize          int
	enabled            bool
	verbose            bool
	lean               bool
	commit             bool
	period             int
	onOperation        bool // TODO Remove in favor of binary search for invariant violation
	allInvariants      bool
	genesisTime        int64
)

// getSimulatorFlags gets the values of all the available simulation flags
func getSimulatorFlags() {
	flag.StringVar(&genesisFile, "Genesis", "", "custom simulation genesis file; cannot be used with params file")
	flag.StringVar(&paramsFile, "Params", "", "custom simulation params file which overrides any random params; cannot be used with genesis")
	flag.StringVar(&exportParamsPath, "ExportParamsPath", "", "custom file path to save the exported params JSON")
	flag.IntVar(&exportParamsHeight, "ExportParamsHeight", 0, "height to which export the randomly generated params")
	flag.StringVar(&exportStatePath, "ExportStatePath", "", "custom file path to save the exported app state JSON")
	flag.StringVar(&exportStatsPath, "ExportStatsPath", "", "custom file path to save the exported simulation statistics JSON")
	flag.Int64Var(&seed, "Seed", 42, "simulation random seed")
	flag.IntVar(&initialBlockHeight, "InitialBlockHeight", 1, "initial block to start the simulation")
	flag.IntVar(&numBlocks, "NumBlocks", 500, "number of new blocks to simulate from the initial block height")
	flag.IntVar(&blockSize, "BlockSize", 200, "operations per block")
	flag.BoolVar(&enabled, "Enabled", false, "enable the simulation")
	flag.BoolVar(&verbose, "Verbose", false, "verbose log output")
	flag.BoolVar(&lean, "Lean", false, "lean simulation log output")
	flag.BoolVar(&commit, "Commit", false, "have the simulation commit")
	flag.IntVar(&period, "Period", 1, "run slow invariants only once every period assertions")
	flag.BoolVar(&onOperation, "SimulateEveryOperation", false, "run slow invariants every operation")
	flag.BoolVar(&allInvariants, "PrintAllInvariants", false, "print all invariants if a broken invariant is found")
	flag.Int64Var(&genesisTime, "GenesisTime", 0, "override genesis UNIX time instead of using a random UNIX time")
}

//---------------------------------------------------------------------
// Simulation Utils

// GetSimulationLog unmarshals the KVPair's Value to the corresponding type based on the
// each's module store key and the prefix bytes of the KVPair's key.
func GetSimulationLog(storeName string, cdcA, cdcB *codec.Codec, kvs []cmn.KVPair) (log string) {
	var kvA, kvB cmn.KVPair
	for i := 0; i < len(kvs); i += 2 {
		kvA = kvs[i]
		kvB = kvs[i+1]

		if len(kvA.Value) == 0 && len(kvB.Value) == 0 {
			// skip if the value doesn't have any bytes
			continue
		}

		switch storeName {
		case auth.StoreKey:
			log += DecodeAccountStore(cdcA, cdcB, kvA, kvB)
		case mint.StoreKey:
			log += DecodeMintStore(cdcA, cdcB, kvA, kvB)
		case staking.StoreKey:
			log += DecodeStakingStore(cdcA, cdcB, kvA, kvB)
		case slashing.StoreKey:
			log += DecodeSlashingStore(cdcA, cdcB, kvA, kvB)
		case gov.StoreKey:
			log += DecodeGovStore(cdcA, cdcB, kvA, kvB)
		case distribution.StoreKey:
			log += DecodeDistributionStore(cdcA, cdcB, kvA, kvB)
		case supply.StoreKey:
			log += DecodeSupplyStore(cdcA, cdcB, kvA, kvB)
		default:
			log += fmt.Sprintf("store A %X => %X\nstore B %X => %X\n", kvA.Key, kvA.Value, kvB.Key, kvB.Value)
		}
	}

	return
}

// DecodeAccountStore unmarshals the KVPair's Value to the corresponding auth type
func DecodeAccountStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], auth.AddressStoreKeyPrefix):
		var accA, accB auth.Account
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &accA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &accB)
		return fmt.Sprintf("%v\n%v", accA, accB)
	case bytes.Equal(kvA.Key, auth.GlobalAccountNumberKey):
		var globalAccNumberA, globalAccNumberB uint64
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &globalAccNumberA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &globalAccNumberB)
		return fmt.Sprintf("GlobalAccNumberA: %d\nGlobalAccNumberB: %d", globalAccNumberA, globalAccNumberB)
	default:
		panic(fmt.Sprintf("invalid account key %X", kvA.Key))
	}
}

// DecodeMintStore unmarshals the KVPair's Value to the corresponding mint type
func DecodeMintStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key, mint.MinterKey):
		var minterA, minterB mint.Minter
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &minterA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &minterB)
		return fmt.Sprintf("%v\n%v", minterA, minterB)
	default:
		panic(fmt.Sprintf("invalid mint key %X", kvA.Key))
	}
}

// DecodeDistributionStore unmarshals the KVPair's Value to the corresponding distribution type
func DecodeDistributionStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], distribution.FeePoolKey):
		var feePoolA, feePoolB distribution.FeePool
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &feePoolA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &feePoolB)
		return fmt.Sprintf("%v\n%v", feePoolA, feePoolB)

	case bytes.Equal(kvA.Key[:1], distribution.ProposerKey):
		return fmt.Sprintf("%v\n%v", sdk.ConsAddress(kvA.Value), sdk.ConsAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorOutstandingRewardsPrefix):
		var rewardsA, rewardsB distribution.ValidatorOutstandingRewards
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], distribution.DelegatorWithdrawAddrPrefix):
		return fmt.Sprintf("%v\n%v", sdk.AccAddress(kvA.Value), sdk.AccAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], distribution.DelegatorStartingInfoPrefix):
		var infoA, infoB distribution.DelegatorStartingInfo
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &infoA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &infoB)
		return fmt.Sprintf("%v\n%v", infoA, infoB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorHistoricalRewardsPrefix):
		var rewardsA, rewardsB distribution.ValidatorHistoricalRewards
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorCurrentRewardsPrefix):
		var rewardsA, rewardsB distribution.ValidatorCurrentRewards
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorAccumulatedCommissionPrefix):
		var commissionA, commissionB distribution.ValidatorAccumulatedCommission
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &commissionA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &commissionB)
		return fmt.Sprintf("%v\n%v", commissionA, commissionB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorSlashEventPrefix):
		var eventA, eventB distribution.ValidatorSlashEvent
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &eventA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &eventB)
		return fmt.Sprintf("%v\n%v", eventA, eventB)

	default:
		panic(fmt.Sprintf("invalid distribution key prefix %X", kvA.Key[:1]))
	}
}

// DecodeStakingStore unmarshals the KVPair's Value to the corresponding staking type
func DecodeStakingStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], staking.LastTotalPowerKey):
		var powerA, powerB sdk.Int
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &powerA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &powerB)
		return fmt.Sprintf("%v\n%v", powerA, powerB)

	case bytes.Equal(kvA.Key[:1], staking.ValidatorsKey):
		var validatorA, validatorB staking.Validator
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &validatorA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &validatorB)
		return fmt.Sprintf("%v\n%v", validatorA, validatorB)

	case bytes.Equal(kvA.Key[:1], staking.LastValidatorPowerKey),
		bytes.Equal(kvA.Key[:1], staking.ValidatorsByConsAddrKey),
		bytes.Equal(kvA.Key[:1], staking.ValidatorsByPowerIndexKey):
		return fmt.Sprintf("%v\n%v", sdk.ValAddress(kvA.Value), sdk.ValAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], staking.DelegationKey):
		var delegationA, delegationB staking.Delegation
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &delegationA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &delegationB)
		return fmt.Sprintf("%v\n%v", delegationA, delegationB)

	case bytes.Equal(kvA.Key[:1], staking.UnbondingDelegationKey),
		bytes.Equal(kvA.Key[:1], staking.UnbondingDelegationByValIndexKey):
		var ubdA, ubdB staking.UnbondingDelegation
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &ubdA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &ubdB)
		return fmt.Sprintf("%v\n%v", ubdA, ubdB)

	case bytes.Equal(kvA.Key[:1], staking.RedelegationKey),
		bytes.Equal(kvA.Key[:1], staking.RedelegationByValSrcIndexKey):
		var redA, redB staking.Redelegation
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &redA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &redB)
		return fmt.Sprintf("%v\n%v", redA, redB)

	default:
		panic(fmt.Sprintf("invalid staking key prefix %X", kvA.Key[:1]))
	}
}

// DecodeSlashingStore unmarshals the KVPair's Value to the corresponding slashing type
func DecodeSlashingStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], slashing.ValidatorSigningInfoKey):
		var infoA, infoB slashing.ValidatorSigningInfo
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &infoA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &infoB)
		return fmt.Sprintf("%v\n%v", infoA, infoB)

	case bytes.Equal(kvA.Key[:1], slashing.ValidatorMissedBlockBitArrayKey):
		var missedA, missedB bool
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &missedA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &missedB)
		return fmt.Sprintf("missedA: %v\nmissedB: %v", missedA, missedB)

	case bytes.Equal(kvA.Key[:1], slashing.AddrPubkeyRelationKey):
		var pubKeyA, pubKeyB crypto.PubKey
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &pubKeyA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &pubKeyB)
		bechPKA := sdk.MustBech32ifyAccPub(pubKeyA)
		bechPKB := sdk.MustBech32ifyAccPub(pubKeyB)
		return fmt.Sprintf("PubKeyA: %s\nPubKeyB: %s", bechPKA, bechPKB)

	default:
		panic(fmt.Sprintf("invalid slashing key prefix %X", kvA.Key[:1]))
	}
}

// DecodeGovStore unmarshals the KVPair's Value to the corresponding gov type
func DecodeGovStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], gov.ProposalsKeyPrefix):
		var proposalA, proposalB gov.Proposal
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &proposalA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &proposalB)
		return fmt.Sprintf("%v\n%v", proposalA, proposalB)

	case bytes.Equal(kvA.Key[:1], gov.ActiveProposalQueuePrefix),
		bytes.Equal(kvA.Key[:1], gov.InactiveProposalQueuePrefix),
		bytes.Equal(kvA.Key[:1], gov.ProposalIDKey):
		proposalIDA := binary.LittleEndian.Uint64(kvA.Value)
		proposalIDB := binary.LittleEndian.Uint64(kvB.Value)
		return fmt.Sprintf("proposalIDA: %d\nProposalIDB: %d", proposalIDA, proposalIDB)

	case bytes.Equal(kvA.Key[:1], gov.DepositsKeyPrefix):
		var depositA, depositB gov.Deposit
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &depositA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &depositB)
		return fmt.Sprintf("%v\n%v", depositA, depositB)

	case bytes.Equal(kvA.Key[:1], gov.VotesKeyPrefix):
		var voteA, voteB gov.Vote
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &voteA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &voteB)
		return fmt.Sprintf("%v\n%v", voteA, voteB)

	default:
		panic(fmt.Sprintf("invalid governance key prefix %X", kvA.Key[:1]))
	}
}

// DecodeSupplyStore unmarshals the KVPair's Value to the corresponding supply type
func DecodeSupplyStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], supply.SupplyKey):
		var supplyA, supplyB supply.Supply
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &supplyA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &supplyB)
		return fmt.Sprintf("%v\n%v", supplyB, supplyB)
	default:
		panic(fmt.Sprintf("invalid supply key %X", kvA.Key))
	}
}
