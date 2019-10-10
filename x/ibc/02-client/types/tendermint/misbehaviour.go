package tendermint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CheckMisbehaviour checks if the evidence provided is a misbehaviour
func CheckMisbehaviour(evidence Evidence) sdk.Error {
	// TODO: check evidence
	return nil
}
