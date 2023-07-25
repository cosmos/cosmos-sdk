package types

import (
	"cosmossdk.io/math"
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"
)

type WinningGrants []WinningGrant

type WinningGrant struct {
	DAO            sdk.AccAddress `json:"dao"`
	Amount         math.Int       `json:"amount"`
	ExpireAtHeight math.Int       `json:"expire_at_height"`
	YesRatio       math.LegacyDec `json:"yes_ratio"`
	ProposalID     math.Int       `json:"proposal_id"`
	MaxCap         math.Int       `json:"max_cap"`
}

// RawWinningGrant is an intermediate type to help unmarshal JSON
type RawWinningGrant struct {
	DAO            string          `json:"dao"`
	Amount         json.RawMessage `json:"amount"`
	ExpireAtHeight json.RawMessage `json:"expire_at_height"`
	YesRatio       string          `json:"yes_ratio"`
	ProposalID     json.RawMessage `json:"proposal_id"`
	MaxCap         json.RawMessage `json:"max_cap"`
}

func (wg *WinningGrant) UnmarshalJSON(b []byte) error {
	var raw RawWinningGrant
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	wg.DAO, err = sdk.AccAddressFromBech32(raw.DAO)
	if err != nil {
		return err
	}

	wg.Amount, err = parseRawInt(raw.Amount)
	if err != nil {
		return err
	}

	wg.ExpireAtHeight, err = parseRawInt(raw.ExpireAtHeight)
	if err != nil {
		return err
	}

	wg.MaxCap, err = parseRawInt(raw.MaxCap)
	if err != nil {
		return err
	}

	yesRatio, err := math.LegacyNewDecFromStr(raw.YesRatio)
	if err != nil {
		return err
	}
	wg.YesRatio = yesRatio

	wg.ProposalID, err = parseRawInt(raw.ProposalID)
	if err != nil {
		return err
	}

	return nil
}

func parseRawInt(raw json.RawMessage) (math.Int, error) {
	var i int64
	var s string
	var err error

	if err = json.Unmarshal(raw, &i); err == nil {
		return math.NewInt(i), nil
	}

	if err = json.Unmarshal(raw, &s); err != nil {
		return math.Int{}, err
	}

	if i, err = strconv.ParseInt(s, 10, 64); err != nil {
		return math.Int{}, err
	}

	return math.NewInt(i), nil
}
