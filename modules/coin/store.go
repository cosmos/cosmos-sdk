package coin

import (
	"fmt"

	wire "github.com/tendermint/go-wire"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/state"
)

// GetAccount - Get account from store and address
func GetAccount(store state.SimpleDB, addr sdk.Actor) (Account, error) {
	// if the actor is another chain, we use one address for the chain....
	addr = ChainAddr(addr)
	acct, err := loadAccount(store, addr.Bytes())

	// for empty accounts, don't return an error, but rather an empty account
	if IsNoAccountErr(err) {
		err = nil
	}
	return acct, err
}

// CheckCoins makes sure there are funds, but doesn't change anything
func CheckCoins(store state.SimpleDB, addr sdk.Actor, coins Coins) (Coins, error) {
	// if the actor is another chain, we use one address for the chain....
	addr = ChainAddr(addr)

	acct, err := updateCoins(store, addr, coins)
	return acct.Coins, err
}

// ChangeCoins changes the money, returns error if it would be negative
func ChangeCoins(store state.SimpleDB, addr sdk.Actor, coins Coins) (Coins, error) {
	// if the actor is another chain, we use one address for the chain....
	addr = ChainAddr(addr)

	acct, err := updateCoins(store, addr, coins)
	if err != nil {
		return acct.Coins, err
	}

	err = storeAccount(store, addr.Bytes(), acct)
	return acct.Coins, err
}

// ChainAddr collapses all addresses from another chain into one, so we can
// keep an over-all balance
//
// TODO: is there a better way to do this?
func ChainAddr(addr sdk.Actor) sdk.Actor {
	if addr.ChainID == "" {
		return addr
	}
	addr.App = ""
	addr.Address = nil
	return addr
}

// updateCoins will load the account, make all checks, and return the updated account.
//
// it doesn't save anything, that is up to you to decide (Check/Change Coins)
func updateCoins(store state.SimpleDB, addr sdk.Actor, coins Coins) (acct Account, err error) {
	acct, err = loadAccount(store, addr.Bytes())
	// we can increase an empty account...
	if IsNoAccountErr(err) && coins.IsPositive() {
		err = nil
	}
	if err != nil {
		return acct, err
	}

	// check amount
	final := acct.Coins.Plus(coins)
	if !final.IsNonnegative() {
		return acct, ErrInsufficientFunds()
	}

	acct.Coins = final
	return acct, nil
}

// Account - coin account structure
type Account struct {
	// Coins is how much is on the account
	Coins Coins `json:"coins"`
	// Credit is how much has been "fronted" to the account
	// (this is usually 0 except for trusted chains)
	Credit Coins `json:"credit"`
}

func loadAccount(store state.SimpleDB, key []byte) (acct Account, err error) {
	// fmt.Printf("load:  %X\n", key)
	data := store.Get(key)
	if len(data) == 0 {
		return acct, ErrNoAccount()
	}
	err = wire.ReadBinaryBytes(data, &acct)
	if err != nil {
		msg := fmt.Sprintf("Error reading account %X", key)
		return acct, errors.ErrInternal(msg)
	}
	return acct, nil
}

func storeAccount(store state.SimpleDB, key []byte, acct Account) error {
	// fmt.Printf("store: %X\n", key)
	bin := wire.BinaryBytes(acct)
	store.Set(key, bin)
	return nil // real stores can return error...
}

// HandlerInfo - this is global info on the coin handler
type HandlerInfo struct {
	Issuer sdk.Actor `json:"issuer"`
}

// TODO: where to store these special pieces??
var handlerKey = []byte{12, 34}

func loadHandlerInfo(store state.KVStore) (info HandlerInfo, err error) {
	data := store.Get(handlerKey)
	if len(data) == 0 {
		return info, nil
	}
	err = wire.ReadBinaryBytes(data, &info)
	if err != nil {
		msg := "Error reading handler info"
		return info, errors.ErrInternal(msg)
	}
	return info, nil
}

func storeIssuer(store state.KVStore, issuer sdk.Actor) error {
	info, err := loadHandlerInfo(store)
	if err != nil {
		return err
	}
	info.Issuer = issuer
	d := wire.BinaryBytes(info)
	store.Set(handlerKey, d)
	return nil // real stores can return error...
}
