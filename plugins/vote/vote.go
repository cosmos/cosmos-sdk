package vote

import (
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
	tmsp "github.com/tendermint/tmsp/types"
)

type Vote struct {
	bb *ballotBox
}

type ballotBox struct {
	issue    string
	votesYes int
	votesNo  int
}

type Tx struct {
	voteYes bool
}

func NewVoteInstance(issue string) Vote {
	return Vote{
		&ballotBox{
			issue:    issue,
			votesYes: 0,
			votesNo:  0,
		},
	}
}

func (app Vote) SetOption(store types.KVStore, key string, value string) (log string) {
	return ""
}

//because no coins are being exchanged ctx is unused
func (app Vote) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res tmsp.Result) {

	// Decode tx
	var tx Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return tmsp.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

	//Read the ballotBox from the store
	kvBytes := store.Get([]byte(app.bb.issue))
	var tempBB ballotBox

	//does the issue already exist?
	if kvBytes != nil {
		err := wire.ReadBinaryBytes(kvBytes, &tempBB)
		if err != nil {
			return tmsp.ErrBaseEncodingError.AppendLog("Error decoding BallotBox: " + err.Error())
		}
	} else {

		//TODO add extra fee for opening new issue

		tempBB = ballotBox{
			issue:    app.bb.issue,
			votesYes: 0,
			votesNo:  0,
		}
		issueBytes := wire.BinaryBytes(struct{ ballotBox }{tempBB})
		store.Set([]byte(app.bb.issue), issueBytes)
	}

	//Write the updated ballotBox to the store
	if tx.voteYes {
		tempBB.votesYes += 1
	} else {
		tempBB.votesNo += 1
	}
	issueBytes := wire.BinaryBytes(struct{ ballotBox }{tempBB})
	store.Set([]byte(app.bb.issue), issueBytes)

	return tmsp.OK
}

//unused
func (app Vote) InitChain(store types.KVStore, vals []*tmsp.Validator) {
}

func (app Vote) BeginBlock(store types.KVStore, height uint64) {
}

func (app Vote) EndBlock(store types.KVStore, height uint64) []*tmsp.Validator {
	var diffs []*tmsp.Validator
	return diffs
}
