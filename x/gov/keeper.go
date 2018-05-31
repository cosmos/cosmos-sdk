package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"

	stake "github.com/cosmos/cosmos-sdk/x/stake"
)

var (
	NewProposalIDKey = []byte{0x00} //
	ProposalQueueKey = []byte{0x01} //
	DepositQueueKey  = []byte{0x02} //
	ProposalTypes    = []string{"TextProposal"}
)

type Keeper struct {
	// The reference to the CoinKeeper to modify balances
	ck bank.Keeper

	// The reference to the StakeMapper to get information about stakers
	sm stake.Keeper

	// The (unexposed) keys used to access the stores from the Context.
	storeKey sdk.StoreKey

	// The wire codec for binary encoding/decoding.
	cdc *wire.Codec
}

// NewGovernanceMapper returns a mapper that uses go-wire to (binary) encode and decode gov types.
func NewKeeper(key sdk.StoreKey, ck bank.Keeper, sk stake.Keeper) Keeper {
	cdc := wire.NewCodec()
	return Keeper{
		storeKey: key,
		ck:       ck,
		cdc:      cdc,
		sm:       sk,
	}
}

// Returns the go-wire codec.
func (keeper Keeper) WireCodec() *wire.Codec {
	return keeper.cdc
}

// Implements sdk.AccountMapper.
func (keeper Keeper) SetProposal(ctx sdk.Context, proposal Proposal) {
	store := ctx.KVStore(keeper.storeKey)

	bz, err := keeper.cdc.MarshalBinary(proposal)
	if err != nil {
		panic(err)
	}

	key, _ := keeper.cdc.MarshalBinary(proposal.ProposalID)

	store.Set(key, bz)
}

func (keeper Keeper) GetProposal(ctx sdk.Context, proposalID int64) *Proposal {
	store := ctx.KVStore(keeper.storeKey)
	key, _ := keeper.cdc.MarshalBinary(proposalID)
	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	proposal := &Proposal{}
	err := keeper.cdc.UnmarshalBinary(bz, proposal)
	if err != nil {
		panic(err)
	}

	return proposal
}

func (keeper Keeper) getNewProposalID(ctx sdk.Context) int64 {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(NewProposalIDKey)

	proposalID := new(int64)
	if bz == nil {
		bz, _ = keeper.cdc.MarshalBinary(int64(0))
	}

	err := keeper.cdc.UnmarshalBinary(bz, proposalID) // TODO: switch to UnmarshalBinaryBare when new go-amino gets added
	if err != nil {
		panic("should not happen")
	}

	bz, err = keeper.cdc.MarshalBinary(*proposalID + 1) // TODO: switch to MarshalBinaryBare when new go-amino gets added
	if err != nil {
		panic("should not happen")
	}

	store.Set(NewProposalIDKey, bz)

	return *proposalID
}

func (keeper Keeper) getProposalQueue(ctx sdk.Context) *Queue{
	store := ctx.KVStore(keeper.storeKey)
	queue := GetQueue(keeper,store,ProposalQueueKey)
	return queue
}

func (keeper Keeper) getDepositQueue(ctx sdk.Context) *Queue{
	store := ctx.KVStore(keeper.storeKey)
	queue := GetQueue(keeper,store, DepositQueueKey)
	return queue
}

func (keeper Keeper) popExpiredProposal(ctx sdk.Context) (list []*Proposal){
	depositQueue := keeper.getDepositQueue(ctx)
	for _,proposalID := range depositQueue.value {
		proposal := keeper.GetProposal(ctx,proposalID)
		if proposal.isDepositPeriodOver(ctx.BlockHeight()){
			list = append(list,proposal)
		}else {
			break
		}
	}

	for _,proposal := range list{
		if !depositQueue.Remove(proposal.ProposalID) {
			panic("should not happen")
		}
	}
	return list
}

func (keeper Keeper) GetActiveProcedure() *Procedure { // TODO: move to param store and allow for updating of this
	return &Procedure{
		VotingPeriod:      200,
		MinDeposit:        sdk.Coins{{stake.StakingToken, 2}},
		ProposalTypes:     ProposalTypes,
		Threshold:         sdk.NewRat(1, 2),
		Veto:              sdk.NewRat(1, 3),
		FastPassThreshold: sdk.NewRat(2, 3),
		MaxDepositPeriod:  200,
		GovernancePenalty: sdk.NewRat(1, 100),
	}
}

func (keeper Keeper) activateVotingPeriod(ctx sdk.Context, proposal *Proposal) {
	proposal.VotingStartBlock = ctx.BlockHeight()

	pool := keeper.sm.GetPool(ctx)
	proposal.TotalVotingPower = pool.BondedPool

	validatorList := keeper.sm.GetValidators(ctx)
	for _, validator := range validatorList {

		validatorGovInfo := ValidatorGovInfo{
			ProposalID:      proposal.ProposalID,
			ValidatorAddr:   validator.Address,
			InitVotingPower: validator.Power.Evaluate(),
			Minus:           0,
			LastVoteWeight:  -1,
		}

		proposal.ValidatorGovInfos = append(proposal.ValidatorGovInfos, validatorGovInfo)
	}

	keeper.getDepositQueue(ctx).Remove(proposal.ProposalID)
	keeper.getProposalQueue(ctx).Push(proposal.ProposalID)
}
