package stake

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

// name to idetify transaction types
const MsgType = "stake"

// XXX remove: think it makes more sense belonging with the Params so we can
// initialize at genesis - to allow for the same tests we should should make
// the ValidateBasic() function a return from an initializable function
// ValidateBasic(bondDenom string) function
const StakingToken = "fermion"

//Verify interface at compile time
var _, _, _, _ sdk.Msg = &MsgDeclareCandidacy{}, &MsgEditCandidacy{}, &MsgDelegate{}, &MsgUnbond{}

//______________________________________________________________________

// MsgDeclareCandidacy - struct for unbonding transactions
type MsgDeclareCandidacy struct {
	Description
	CandidateAddr sdk.Address   `json:"address"`
	PubKey        crypto.PubKey `json:"pubkey"`
	Bond          sdk.Coin      `json:"bond"`
}

func NewMsgDeclareCandidacy(candidateAddr sdk.Address, pubkey crypto.PubKey,
	bond sdk.Coin, description Description) MsgDeclareCandidacy {
	return MsgDeclareCandidacy{
		Description:   description,
		CandidateAddr: candidateAddr,
		PubKey:        pubkey,
		Bond:          bond,
	}
}

//nolint
func (msg MsgDeclareCandidacy) Type() string                            { return MsgType } //TODO update "stake/declarecandidacy"
func (msg MsgDeclareCandidacy) Get(key interface{}) (value interface{}) { return nil }
func (msg MsgDeclareCandidacy) GetSigners() []sdk.Address               { return []sdk.Address{msg.CandidateAddr} }
func (msg MsgDeclareCandidacy) String() string {
	return fmt.Sprintf("CandidateAddr{Address: %v}", msg.CandidateAddr) // XXX fix
}

// get the bytes for the message signer to sign on
func (msg MsgDeclareCandidacy) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg MsgDeclareCandidacy) ValidateBasic() sdk.Error {
	if msg.CandidateAddr == nil {
		return ErrCandidateEmpty()
	}
	if msg.Bond.Denom != StakingToken {
		return ErrBadBondingDenom()
	}
	if msg.Bond.Amount <= 0 {
		return ErrBadBondingAmount()
		// return sdk.ErrInvalidCoins(sdk.Coins{msg.Bond}.String())
	}
	empty := Description{}
	if msg.Description == empty {
		return newError(CodeInvalidInput, "description must be included")
	}
	return nil
}

//______________________________________________________________________

// MsgEditCandidacy - struct for editing a candidate
type MsgEditCandidacy struct {
	Description
	CandidateAddr sdk.Address `json:"address"`
}

func NewMsgEditCandidacy(candidateAddr sdk.Address, description Description) MsgEditCandidacy {
	return MsgEditCandidacy{
		Description:   description,
		CandidateAddr: candidateAddr,
	}
}

//nolint
func (msg MsgEditCandidacy) Type() string                            { return MsgType } //TODO update "stake/msgeditcandidacy"
func (msg MsgEditCandidacy) Get(key interface{}) (value interface{}) { return nil }
func (msg MsgEditCandidacy) GetSigners() []sdk.Address               { return []sdk.Address{msg.CandidateAddr} }
func (msg MsgEditCandidacy) String() string {
	return fmt.Sprintf("CandidateAddr{Address: %v}", msg.CandidateAddr) // XXX fix
}

// get the bytes for the message signer to sign on
func (msg MsgEditCandidacy) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg MsgEditCandidacy) ValidateBasic() sdk.Error {
	if msg.CandidateAddr == nil {
		return ErrCandidateEmpty()
	}
	empty := Description{}
	if msg.Description == empty {
		return newError(CodeInvalidInput, "Transaction must include some information to modify")
	}
	return nil
}

//______________________________________________________________________

// MsgDelegate - struct for bonding transactions
type MsgDelegate struct {
	DelegatorAddr sdk.Address `json:"address"`
	CandidateAddr sdk.Address `json:"address"`
	Bond          sdk.Coin    `json:"bond"`
}

func NewMsgDelegate(delegatorAddr, candidateAddr sdk.Address, bond sdk.Coin) MsgDelegate {
	return MsgDelegate{
		DelegatorAddr: delegatorAddr,
		CandidateAddr: candidateAddr,
		Bond:          bond,
	}
}

//nolint
func (msg MsgDelegate) Type() string                            { return MsgType } //TODO update "stake/msgeditcandidacy"
func (msg MsgDelegate) Get(key interface{}) (value interface{}) { return nil }
func (msg MsgDelegate) GetSigners() []sdk.Address               { return []sdk.Address{msg.DelegatorAddr} }
func (msg MsgDelegate) String() string {
	return fmt.Sprintf("Addr{Address: %v}", msg.DelegatorAddr) // XXX fix
}

// get the bytes for the message signer to sign on
func (msg MsgDelegate) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg MsgDelegate) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr == nil {
		return ErrBadDelegatorAddr()
	}
	if msg.CandidateAddr == nil {
		return ErrBadCandidateAddr()
	}
	if msg.Bond.Denom != StakingToken {
		return ErrBadBondingDenom()
	}
	if msg.Bond.Amount <= 0 {
		return ErrBadBondingAmount()
		// return sdk.ErrInvalidCoins(sdk.Coins{msg.Bond}.String())
	}
	return nil
}

//______________________________________________________________________

// MsgUnbond - struct for unbonding transactions
type MsgUnbond struct {
	DelegatorAddr sdk.Address `json:"address"`
	CandidateAddr sdk.Address `json:"address"`
	Shares        string      `json:"shares"`
}

func NewMsgUnbond(delegatorAddr, candidateAddr sdk.Address, shares string) MsgUnbond {
	return MsgUnbond{
		DelegatorAddr: delegatorAddr,
		CandidateAddr: candidateAddr,
		Shares:        shares,
	}
}

//nolint
func (msg MsgUnbond) Type() string                            { return MsgType } //TODO update "stake/msgeditcandidacy"
func (msg MsgUnbond) Get(key interface{}) (value interface{}) { return nil }
func (msg MsgUnbond) GetSigners() []sdk.Address               { return []sdk.Address{msg.DelegatorAddr} }
func (msg MsgUnbond) String() string {
	return fmt.Sprintf("Addr{Address: %v}", msg.DelegatorAddr) // XXX fix
}

// get the bytes for the message signer to sign on
func (msg MsgUnbond) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg MsgUnbond) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr == nil {
		return ErrBadDelegatorAddr()
	}
	if msg.CandidateAddr == nil {
		return ErrBadCandidateAddr()
	}
	if msg.Shares != "MAX" {
		rat, err := sdk.NewRatFromDecimal(msg.Shares)
		if err != nil {
			return ErrBadShares()
		}
		if rat.IsZero() || rat.LT(sdk.ZeroRat) {
			return ErrBadShares()
		}
	}
	return nil
}
