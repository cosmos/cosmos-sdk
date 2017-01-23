package paytovote

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
)

const (
	TypeByteCreateIssue byte = 0x00
	TypeByteVoteFor     byte = 0x01
	TypeByteVoteAgainst byte = 0x02
	TypeByteVoteSpoiled byte = 0x03
)

type P2VPluginState struct {
	TotalCost    types.Coins
	Issue        string
	votesFor     int
	votesAgainst int
	votesSpoiled int
}

type P2VTx struct {
	Valid            bool
	Issue            string      //Issue being voted for
	ActionTypeByte   byte        //How is the vote being cast
	Cost2Vote        types.Coins //Cost to vote
	Cost2CreateIssue types.Coins //Cost to create a new issue
}

//--------------------------------------------------------------------------------

type P2VPlugin struct {
	name string
}

func (p2v *P2VPlugin) Name() string {
	return p2v.name
}

func (p2v *P2VPlugin) StateKey(issue string) []byte {
	return []byte(fmt.Sprintf("P2VPlugin{name=%v,issue=%v}.State", p2v.name, issue))
}

func New(name string) *P2VPlugin {
	return &P2VPlugin{
		name: name,
	}
}

func newState(issue string) P2VPluginState {
	return P2VPluginState{
		TotalCost:    types.Coins{},
		Issue:        issue,
		votesFor:     0,
		votesAgainst: 0,
		votesSpoiled: 0,
	}
}

func (p2v *P2VPlugin) SetOption(store types.KVStore, key string, value string) (log string) {
	return ""
}

func (p2v *P2VPlugin) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res abci.Result) {

	// Decode tx
	var tx P2VTx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

	// Validate tx
	if !tx.Valid {
		return abci.ErrInternalError.AppendLog("P2VTx.Valid must be true")
	}

	if len(tx.Issue) == 0 {
		return abci.ErrInternalError.AppendLog("P2VTx.Issue must have a length greater than 0")
	}

	if !tx.Cost2Vote.IsValid() {
		return abci.ErrInternalError.AppendLog("P2VTx.Cost2Vote is not sorted or has zero amounts")
	}

	if !tx.Cost2Vote.IsNonnegative() {
		return abci.ErrInternalError.AppendLog("P2VTx.Cost2Vote must be nonnegative")
	}

	if !tx.Cost2CreateIssue.IsValid() {
		return abci.ErrInternalError.AppendLog("P2VTx.Cost2CreateIssue is not sorted or has zero amounts")
	}

	if !tx.Cost2CreateIssue.IsNonnegative() {
		return abci.ErrInternalError.AppendLog("P2VTx.Cost2CreateIssue must be nonnegative")
	}

	// Load P2VPluginState
	var p2vState P2VPluginState
	p2vStateBytes := store.Get(p2v.StateKey(tx.Issue))

	//Determine if the issue already exists
	issueExists := true

	if len(p2vStateBytes) > 0 { //is there a record of the issue existing?
		err = wire.ReadBinaryBytes(p2vStateBytes, &p2vState)
		if err != nil {
			return abci.ErrInternalError.AppendLog("Error decoding state: " + err.Error())
		}
	} else {
		issueExists = false
	}

	returnLeftover := func(cost types.Coins) {
		leftoverCoins := ctx.Coins.Minus(cost)
		if !leftoverCoins.IsZero() {
			// TODO If there are any funds left over, return funds.
			// ctx.CallerAccount is synced w/ store, so just modify that and store it.
		}
	}

	switch {
	case tx.ActionTypeByte == TypeByteCreateIssue && issueExists:
		return abci.ErrInsufficientFunds.AppendLog("Cannot create an already existing issue")
	case tx.ActionTypeByte != TypeByteCreateIssue && !issueExists:
		return abci.ErrInsufficientFunds.AppendLog("Tx Issue not found")
	case tx.ActionTypeByte == TypeByteCreateIssue && !issueExists:
		// Did the caller provide enough coins?
		if !ctx.Coins.IsGTE(tx.Cost2CreateIssue) {
			return abci.ErrInsufficientFunds.AppendLog("Tx Funds insufficient for creating a new issue")
		}

		// Update P2VPluginState
		newP2VState := newState(tx.Issue)
		newP2VState.TotalCost = newP2VState.TotalCost.Plus(tx.Cost2Vote)

		// Save P2VPluginState
		store.Set(p2v.StateKey(tx.Issue), wire.BinaryBytes(newP2VState))

		returnLeftover(tx.Cost2CreateIssue)

	case tx.ActionTypeByte != TypeByteCreateIssue && issueExists:
		// Did the caller provide enough coins?
		if !ctx.Coins.IsGTE(tx.Cost2Vote) {
			return abci.ErrInsufficientFunds.AppendLog("Tx Funds insufficient for voting")
		}

		switch tx.ActionTypeByte {
		case TypeByteVoteFor:
			p2vState.votesFor += 1
		case TypeByteVoteAgainst:
			p2vState.votesAgainst += 1
		case TypeByteVoteSpoiled:
			p2vState.votesSpoiled += 1
		default:
			return abci.ErrInternalError.AppendLog("P2VTx.ActionTypeByte was not recognized")
		}

		// Update P2VPluginState
		p2vState.TotalCost = p2vState.TotalCost.Plus(tx.Cost2Vote)

		// Save P2VPluginState
		store.Set(p2v.StateKey(tx.Issue), wire.BinaryBytes(p2vState))

		returnLeftover(tx.Cost2CreateIssue)
	}

	return abci.NewResultOK(wire.BinaryBytes(p2vState), "")
}

func (p2v *P2VPlugin) InitChain(store types.KVStore, vals []*abci.Validator) {
}

func (p2v *P2VPlugin) BeginBlock(store types.KVStore, height uint64) {
}

func (p2v *P2VPlugin) EndBlock(store types.KVStore, height uint64) []*abci.Validator {
	return nil
}
