package gov

import (
	codec "github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter store default namestore
const (
	DefaultParamspace = "gov"
)

// Parameter store key
var (
	ParamStoreKeyDepositProcedure  = []byte("depositprocedure")
	ParamStoreKeyVotingProcedure   = []byte("votingprocedure")
	ParamStoreKeyTallyingProcedure = []byte("tallyingprocedure")
)

// Type declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		ParamStoreKeyDepositProcedure, DepositProcedure{},
		ParamStoreKeyVotingProcedure, VotingProcedure{},
		ParamStoreKeyTallyingProcedure, TallyingProcedure{},
	)
}

// Governance Keeper
type Keeper struct {
	// The reference to the Param Keeper to get and set Global Params
	paramsKeeper params.Keeper

	// The reference to the Paramstore to get and set gov specific params
	paramSpace params.Subspace

	// The reference to the CoinKeeper to modify balances
	ck bank.Keeper

	// The ValidatorSet to get information about validators
	vs sdk.ValidatorSet

	// The reference to the DelegationSet to get information about delegators
	ds sdk.DelegationSet

	// The (unexposed) keys used to access the stores from the Context.
	storeKey sdk.StoreKey

	// The codec codec for binary encoding/decoding.
	cdc *codec.Codec

	// Reserved codespace
	codespace sdk.CodespaceType
}

// NewKeeper returns a governance keeper. It handles:
// - submitting governance proposals
// - depositing funds into proposals, and activating upon sufficient funds being deposited
// - users voting on proposals, with weight proportional to stake in the system
// - and tallying the result of the vote.
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramsKeeper params.Keeper, paramSpace params.Subspace, ck bank.Keeper, ds sdk.DelegationSet, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		storeKey:     key,
		paramsKeeper: paramsKeeper,
		paramSpace:   paramSpace.WithKeyTable(ParamKeyTable()),
		ck:           ck,
		ds:           ds,
		vs:           ds.GetValidatorSet(),
		cdc:          cdc,
		codespace:    codespace,
	}
}

// =====================================================
// Proposals

// Creates a NewProposal
func (keeper Keeper) NewTextProposal(ctx sdk.Context, title string, description string, proposalType ProposalKind) (proposalID int64, err error) {
	proposalID, err = keeper.getNewProposalID(ctx)
	if err != nil {
		return
	}
	proposal := TextProposal{
		Abstract: ProposalAbstract{
			Title:       title,
			Description: description,
		},

		ProposalType: proposalType,
	}
	info := ProposalInfo{
		ProposalID:   proposalID,
		Status:       StatusDepositPeriod,
		TallyResult:  EmptyTallyResult(),
		TotalDeposit: sdk.Coins{},
		SubmitTime:   ctx.BlockHeader().Time,
	}

	keeper.SetProposal(ctx, proposalID, proposal)
	keeper.InactiveInfoQueuePush(ctx, info)
	return
}

// Get Proposal from store by ProposalID
func (keeper Keeper) GetProposalInfo(ctx sdk.Context, proposalID int64) (info ProposalInfo) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyInfo(proposalID))
	if bz == nil {
		return
	}

	keeper.cdc.MustUnmarshalBinary(bz, &info)

	return
}

// Implements sdk.AccountKeeper.
func (keeper Keeper) SetProposalInfo(ctx sdk.Context, info ProposalInfo) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinary(info)
	store.Set(KeyInfo(info.ProposalID), bz)
}

// Implements sdk.AccountKeeper.
func (keeper Keeper) DeleteProposalInfo(ctx sdk.Context, proposalID int64) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(KeyInfo(proposalID))
}

// Get Proposal from store by ProposalID
func (keeper Keeper) GetProposal(ctx sdk.Context, proposalID int64) Proposal {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyInfo(proposalID))
	if bz == nil {
		return nil
	}

	var proposal Proposal
	keeper.cdc.MustUnmarshalBinary(bz, &proposal)

	return proposal
}

// Implements sdk.AccountKeeper.
func (keeper Keeper) SetProposal(ctx sdk.Context, proposalID int64, proposal Proposal) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinary(proposal)
	store.Set(KeyInfo(proposalID), bz)
}

// Implements sdk.AccountKeeper.
func (keeper Keeper) DeleteProposal(ctx sdk.Context, proposalID int64) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(KeyInfo(proposalID))
}

// Get Proposal from store by ProposalID
func (keeper Keeper) GetProposalsFiltered(ctx sdk.Context, voterAddr sdk.AccAddress, depositerAddr sdk.AccAddress, status ProposalStatus, numLatest int64) []Proposal {

	maxProposalID, err := keeper.peekCurrentProposalID(ctx)
	if err != nil {
		return nil
	}

	matchingProposals := []Proposal{}

	if numLatest <= 0 {
		numLatest = maxProposalID
	}

	for proposalID := maxProposalID - numLatest; proposalID < maxProposalID; proposalID++ {
		if voterAddr != nil && len(voterAddr) != 0 {
			_, found := keeper.GetVote(ctx, proposalID, voterAddr)
			if !found {
				continue
			}
		}

		if depositerAddr != nil && len(depositerAddr) != 0 {
			_, found := keeper.GetDeposit(ctx, proposalID, depositerAddr)
			if !found {
				continue
			}
		}

		proposal := keeper.GetProposal(ctx, proposalID)
		if proposal == nil {
			continue
		}

		if validProposalStatus(status) {
			if keeper.GetProposalInfo(ctx, proposalID).Status != status {
				continue
			}
		}

		matchingProposals = append(matchingProposals, proposal)
	}
	return matchingProposals
}

func (keeper Keeper) setInitialProposalID(ctx sdk.Context, proposalID int64) sdk.Error {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyNextProposalID)
	if bz != nil {
		return ErrInvalidGenesis(keeper.codespace, "Initial ProposalID already set")
	}
	bz = keeper.cdc.MustMarshalBinary(proposalID)
	store.Set(KeyNextProposalID, bz)
	return nil
}

// Get the last used proposal ID
func (keeper Keeper) GetLastProposalID(ctx sdk.Context) (proposalID int64) {
	proposalID, err := keeper.peekCurrentProposalID(ctx)
	if err != nil {
		return 0
	}
	proposalID--
	return
}

// Gets the next available ProposalID and increments it
func (keeper Keeper) getNewProposalID(ctx sdk.Context) (proposalID int64, err sdk.Error) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyNextProposalID)
	if bz == nil {
		return -1, ErrInvalidGenesis(keeper.codespace, "InitialProposalID never set")
	}
	keeper.cdc.MustUnmarshalBinary(bz, &proposalID)
	bz = keeper.cdc.MustMarshalBinary(proposalID + 1)
	store.Set(KeyNextProposalID, bz)
	return proposalID, nil
}

// Peeks the next available ProposalID without incrementing it
func (keeper Keeper) peekCurrentProposalID(ctx sdk.Context) (proposalID int64, err sdk.Error) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyNextProposalID)
	if bz == nil {
		return -1, ErrInvalidGenesis(keeper.codespace, "InitialProposalID never set")
	}
	keeper.cdc.MustUnmarshalBinary(bz, &proposalID)
	return proposalID, nil
}

func (keeper Keeper) activateVotingPeriod(ctx sdk.Context, info ProposalInfo) {
	info.VotingStartTime = ctx.BlockHeader().Time
	info.Status = StatusVotingPeriod
	keeper.SetProposalInfo(ctx, info)
	keeper.ActiveInfoQueuePush(ctx, info)
}

// =====================================================
// Procedures

// Returns the current Deposit Procedure from the global param store
// nolint: errcheck
func (keeper Keeper) GetDepositProcedure(ctx sdk.Context) DepositProcedure {
	var depositProcedure DepositProcedure
	keeper.paramSpace.Get(ctx, ParamStoreKeyDepositProcedure, &depositProcedure)
	return depositProcedure
}

// Returns the current Voting Procedure from the global param store
// nolint: errcheck
func (keeper Keeper) GetVotingProcedure(ctx sdk.Context) VotingProcedure {
	var votingProcedure VotingProcedure
	keeper.paramSpace.Get(ctx, ParamStoreKeyVotingProcedure, &votingProcedure)
	return votingProcedure
}

// Returns the current Tallying Procedure from the global param store
// nolint: errcheck
func (keeper Keeper) GetTallyingProcedure(ctx sdk.Context) TallyingProcedure {
	var tallyingProcedure TallyingProcedure
	keeper.paramSpace.Get(ctx, ParamStoreKeyTallyingProcedure, &tallyingProcedure)
	return tallyingProcedure
}

// nolint: errcheck
func (keeper Keeper) setDepositProcedure(ctx sdk.Context, depositProcedure DepositProcedure) {
	keeper.paramSpace.Set(ctx, ParamStoreKeyDepositProcedure, &depositProcedure)
}

// nolint: errcheck
func (keeper Keeper) setVotingProcedure(ctx sdk.Context, votingProcedure VotingProcedure) {
	keeper.paramSpace.Set(ctx, ParamStoreKeyVotingProcedure, &votingProcedure)
}

// nolint: errcheck
func (keeper Keeper) setTallyingProcedure(ctx sdk.Context, tallyingProcedure TallyingProcedure) {
	keeper.paramSpace.Set(ctx, ParamStoreKeyTallyingProcedure, &tallyingProcedure)
}

// =====================================================
// Votes

// Adds a vote on a specific proposal
func (keeper Keeper) AddVote(ctx sdk.Context, proposalID int64, voterAddr sdk.AccAddress, option VoteOption) sdk.Error {
	info := keeper.GetProposalInfo(ctx, proposalID)
	if info.IsEmpty() {
		return ErrUnknownProposal(keeper.codespace, proposalID)
	}
	if info.Status != StatusVotingPeriod {
		return ErrInactiveProposal(keeper.codespace, proposalID)
	}

	if !validVoteOption(option) {
		return ErrInvalidVote(keeper.codespace, option)
	}

	vote := Vote{
		ProposalID: proposalID,
		Voter:      voterAddr,
		Option:     option,
	}
	keeper.setVote(ctx, proposalID, voterAddr, vote)

	return nil
}

// Gets the vote of a specific voter on a specific proposal
func (keeper Keeper) GetVote(ctx sdk.Context, proposalID int64, voterAddr sdk.AccAddress) (Vote, bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyVote(proposalID, voterAddr))
	if bz == nil {
		return Vote{}, false
	}
	var vote Vote
	keeper.cdc.MustUnmarshalBinary(bz, &vote)
	return vote, true
}

func (keeper Keeper) setVote(ctx sdk.Context, proposalID int64, voterAddr sdk.AccAddress, vote Vote) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinary(vote)
	store.Set(KeyVote(proposalID, voterAddr), bz)
}

// Gets all the votes on a specific proposal
func (keeper Keeper) GetVotes(ctx sdk.Context, proposalID int64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, KeyVotesSubspace(proposalID))
}

func (keeper Keeper) deleteVote(ctx sdk.Context, proposalID int64, voterAddr sdk.AccAddress) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(KeyVote(proposalID, voterAddr))
}

// =====================================================
// Deposits

// Gets the deposit of a specific depositer on a specific proposal
func (keeper Keeper) GetDeposit(ctx sdk.Context, proposalID int64, depositerAddr sdk.AccAddress) (Deposit, bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyDeposit(proposalID, depositerAddr))
	if bz == nil {
		return Deposit{}, false
	}
	var deposit Deposit
	keeper.cdc.MustUnmarshalBinary(bz, &deposit)
	return deposit, true
}

func (keeper Keeper) setDeposit(ctx sdk.Context, proposalID int64, depositerAddr sdk.AccAddress, deposit Deposit) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinary(deposit)
	store.Set(KeyDeposit(proposalID, depositerAddr), bz)
}

// Adds or updates a deposit of a specific depositer on a specific proposal
// Activates voting period when appropriate
func (keeper Keeper) AddDeposit(ctx sdk.Context, proposalID int64, depositerAddr sdk.AccAddress, depositAmount sdk.Coins) (sdk.Error, bool) {
	// Checks to see if proposal exists
	info := keeper.GetProposalInfo(ctx, proposalID)
	if info.IsEmpty() {
		return ErrUnknownProposal(keeper.codespace, proposalID), false
	}

	// Check if proposal is still depositable
	if (info.Status != StatusDepositPeriod) && (info.Status != StatusVotingPeriod) {
		return ErrAlreadyFinishedProposal(keeper.codespace, proposalID), false
	}

	// Subtract coins from depositer's account
	_, _, err := keeper.ck.SubtractCoins(ctx, depositerAddr, depositAmount)
	if err != nil {
		return err, false
	}

	// Update Proposal
	info.TotalDeposit = info.TotalDeposit.Plus(depositAmount)
	keeper.SetProposalInfo(ctx, info)

	// Check if deposit tipped proposal into voting period
	// Active voting period if so
	activatedVotingPeriod := false
	if info.Status == StatusDepositPeriod && info.TotalDeposit.IsGTE(keeper.GetDepositProcedure(ctx).MinDeposit) {
		keeper.activateVotingPeriod(ctx, info)
		activatedVotingPeriod = true
	}

	// Add or update deposit object
	currDeposit, found := keeper.GetDeposit(ctx, proposalID, depositerAddr)
	if !found {
		newDeposit := Deposit{depositerAddr, proposalID, depositAmount}
		keeper.setDeposit(ctx, proposalID, depositerAddr, newDeposit)
	} else {
		currDeposit.Amount = currDeposit.Amount.Plus(depositAmount)
		keeper.setDeposit(ctx, proposalID, depositerAddr, currDeposit)
	}

	return nil, activatedVotingPeriod
}

// Gets all the deposits on a specific proposal
func (keeper Keeper) GetDeposits(ctx sdk.Context, proposalID int64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, KeyDepositsSubspace(proposalID))
}

// Returns and deletes all the deposits on a specific proposal
func (keeper Keeper) RefundDeposits(ctx sdk.Context, proposalID int64) {
	store := ctx.KVStore(keeper.storeKey)
	depositsIterator := keeper.GetDeposits(ctx, proposalID)

	for ; depositsIterator.Valid(); depositsIterator.Next() {
		deposit := &Deposit{}
		keeper.cdc.MustUnmarshalBinary(depositsIterator.Value(), deposit)

		_, _, err := keeper.ck.AddCoins(ctx, deposit.Depositer, deposit.Amount)
		if err != nil {
			panic("should not happen")
		}

		store.Delete(depositsIterator.Key())
	}

	depositsIterator.Close()
}

// Deletes all the deposits on a specific proposal without refunding them
func (keeper Keeper) DeleteDeposits(ctx sdk.Context, proposalID int64) {
	store := ctx.KVStore(keeper.storeKey)
	depositsIterator := keeper.GetDeposits(ctx, proposalID)

	for ; depositsIterator.Valid(); depositsIterator.Next() {
		store.Delete(depositsIterator.Key())
	}

	depositsIterator.Close()
}

// =====================================================
// InfoQueues

func (keeper Keeper) getActiveInfoQueue(ctx sdk.Context) InfoQueue {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyActiveInfoQueue)
	if bz == nil {
		return nil
	}

	var infoQueue InfoQueue
	keeper.cdc.MustUnmarshalBinary(bz, &infoQueue)

	return infoQueue
}

func (keeper Keeper) setActiveInfoQueue(ctx sdk.Context, infoQueue InfoQueue) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinary(infoQueue)
	store.Set(KeyActiveInfoQueue, bz)
}

// Return the Proposal at the front of the InfoQueue
func (keeper Keeper) ActiveInfoQueuePeek(ctx sdk.Context) ProposalInfo {
	infoQueue := keeper.getActiveInfoQueue(ctx)
	if len(infoQueue) == 0 {
		return ProposalInfo{}
	}
	return keeper.GetProposalInfo(ctx, infoQueue[0])
}

// Remove and return a Proposal from the front of the InfoQueue
func (keeper Keeper) ActiveInfoQueuePop(ctx sdk.Context) ProposalInfo {
	infoQueue := keeper.getActiveInfoQueue(ctx)
	if len(infoQueue) == 0 {
		return ProposalInfo{}
	}
	frontElement, infoQueue := infoQueue[0], infoQueue[1:]
	keeper.setActiveInfoQueue(ctx, infoQueue)
	return keeper.GetProposalInfo(ctx, frontElement)
}

// Add a infoID to the back of the InfoQueue
func (keeper Keeper) ActiveInfoQueuePush(ctx sdk.Context, info ProposalInfo) {
	infoQueue := append(keeper.getActiveInfoQueue(ctx), info.ProposalID)
	keeper.setActiveInfoQueue(ctx, infoQueue)
}

func (keeper Keeper) getInactiveInfoQueue(ctx sdk.Context) InfoQueue {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyInactiveInfoQueue)
	if bz == nil {
		return nil
	}

	var infoQueue InfoQueue

	keeper.cdc.MustUnmarshalBinary(bz, &infoQueue)

	return infoQueue
}

func (keeper Keeper) setInactiveInfoQueue(ctx sdk.Context, infoQueue InfoQueue) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinary(infoQueue)
	store.Set(KeyInactiveInfoQueue, bz)
}

// Return the Proposal at the front of the InfoQueue
func (keeper Keeper) InactiveInfoQueuePeek(ctx sdk.Context) ProposalInfo {
	infoQueue := keeper.getInactiveInfoQueue(ctx)
	if len(infoQueue) == 0 {
		return ProposalInfo{}
	}
	return keeper.GetProposalInfo(ctx, infoQueue[0])
}

// Remove and return a Proposal from the front of the InfoQueue
func (keeper Keeper) InactiveInfoQueuePop(ctx sdk.Context) ProposalInfo {
	infoQueue := keeper.getInactiveInfoQueue(ctx)
	if len(infoQueue) == 0 {
		return ProposalInfo{}
	}
	frontElement, infoQueue := infoQueue[0], infoQueue[1:]
	keeper.setInactiveInfoQueue(ctx, infoQueue)
	return keeper.GetProposalInfo(ctx, frontElement)
}

// Add a infoID to the back of the InfoQueue
func (keeper Keeper) InactiveInfoQueuePush(ctx sdk.Context, info ProposalInfo) {
	infoQueue := append(keeper.getInactiveInfoQueue(ctx), info.ProposalID)
	keeper.setInactiveInfoQueue(ctx, infoQueue)
}
