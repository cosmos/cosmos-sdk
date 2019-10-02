package tendermint

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

// CheckMisbehaviour checks if the evidence provided is a misbehaviour
func CheckMisbehaviour(evidence exported.Evidence) error {
	_, ok := evidence.(Evidence)
	if !ok {
		return errors.New("header is not from a tendermint consensus") // TODO: create concrete error
	}

	// TODO: check evidence
	return nil
}
