package simplestake

import (
	"encoding/json"

	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// -------------------------
// BondMsg

type BondMsg struct {
	Address sdk.Address   `json:"address"`
	Stake   sdk.Coin      `json:"coins"`
	PubKey  crypto.PubKey `json:"pub_key"`
}

func NewBondMsg(addr sdk.Address, stake sdk.Coin, pubKey crypto.PubKey) BondMsg {
	return BondMsg{
		Address: addr,
		Stake:   stake,
		PubKey:  pubKey,
	}
}

func (msg BondMsg) Type() string {
	return moduleName
}

func (msg BondMsg) ValidateBasic() sdk.Error {
	if msg.Stake.IsZero() {
		return ErrEmptyStake()
	}

	if msg.PubKey.Empty() {
		return sdk.ErrInvalidPubKey("BondMsg.PubKey must not be empty")
	}

	return nil
}

func (msg BondMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg BondMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg BondMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Address}
}

// -------------------------
// UnbondMsg

type UnbondMsg struct {
	Address sdk.Address `json:"address"`
}

func NewUnbondMsg(addr sdk.Address) UnbondMsg {
	return UnbondMsg{
		Address: addr,
	}
}

func (msg UnbondMsg) Type() string {
	return moduleName
}

func (msg UnbondMsg) ValidateBasic() sdk.Error {
	return nil
}

func (msg UnbondMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg UnbondMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg UnbondMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Address}
}
