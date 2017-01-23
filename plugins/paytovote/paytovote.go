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
	Cost2Vote        types.Coins //Cost to vote
	Cost2CreateIssue types.Coins //Cost to create a new issue
	Issue            string      //Issue being voted for
	ActionTypeByte   byte        //How is the vote being cast
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

func newState(issue) P2VPluginState {
	return P2VPluginState{
		TotalCost:    0,
		Issue:        issue,
		votesFor:     0,
		votesAgainst: 0,
		votesSpoiled: 0,
	}
}

func (cp *P2VPlugin) SetOption(store types.KVStore, key string, value string) (log string) {
	return ""
}

func (cp *P2VPlugin) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res abci.Result) {

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

	checkCost := func(cost types.Coins) {
		if !cost.IsValid() {
			return abci.ErrInternalError.AppendLog("P2VTx.Cost is not sorted or has zero amounts")
		}
		if !cost.IsNonnegative() {
			return abci.ErrInternalError.AppendLog("P2VTx.Cost must be nonnegative")
		}
	}
	checkCost(tx.Cost2Vote)
	checkCost(tx.Cost2CreateIssue)

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

	switch true {
	case tx.ActionTypeByte == TypeByteCreateIssue && issueExists:
		return abci.ErrInsufficientFunds.AppendLog("Cannot create an already existing issue")
	case tx.ActionTypeByte != TypeByteCreateIssue && !issueExists:
		return abci.ErrInsufficientFunds.AppendLog("Tx Issue not found")
	case tx.ActionTypeByte == TypeByteCreateIssue && !issueExists:
		// Did the caller provide enough coins?
		if !ctx.Coins.IsGTE(tx.Cost2CreateIssue) {
			return abci.ErrInsufficientFunds.AppendLog("Tx Funds insufficient for creating a new issue")
		}
		store.Set(p2v.StateKey(tx.Issue), wire.BinaryBytes(newState(tx.Issue)))

		// TODO If there are any funds left over, return funds.
		// e.g. !ctx.Coins.Minus(tx.Cost).IsZero()
		// ctx.CallerAccount is synced w/ store, so just modify that and store it.

	case tx.ActionTypeByte != TypeByteCreateIssue && issueExists:
		// Did the caller provide enough coins?
		if !ctx.Coins.IsGTE(tx.Cost2Vote) {
			return abci.ErrInsufficientFunds.AppendLog("Tx Funds insufficient for voting")
		}

		switch true {
		case tx.ActionTypeByte == TypeByteVoteFor:
			p2vState.votesFor += 1
		case tx.ActionTypeByte == TypeByteVoteAgainst:
			p2vState.votesAgainst += 1
		case tx.ActionTypeByte == TypeByteVoteSpoiled:
			p2vState.votesSpoiled += 1
		default:
			return abci.ErrInternalError.AppendLog("P2VTx.ActionTypeByte was not recognized")
		}
		// Update P2VPluginState
		p2vState.TotalCost = p2vState.TotalCost.Plus(tx.Cost)
		// Save P2VPluginState
		store.Set(p2v.StateKey(tx.Issue), wire.BinaryBytes(p2vState))

		// TODO If there are any funds left over, return funds.
		// e.g. !ctx.Coins.Minus(tx.Cost).IsZero()
		// ctx.CallerAccount is synced w/ store, so just modify that and store it.
	}

	return abci.OK
}

func (cp *P2VPlugin) InitChain(store types.KVStore, vals []*abci.Validator) {
}

func (cp *P2VPlugin) BeginBlock(store types.KVStore, height uint64) {
}

func (cp *P2VPlugin) EndBlock(store types.KVStore, height uint64) []*abci.Validator {
	return nil
}
