package types

import (
	"fmt"
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// An Invariant is a function which tests a particular invariant.
// If the invariant has been broken, it should return an error
// containing a descriptive message about what happened.
// The simulator will then halt and print the logs.
type Invariant func(ctx Context) error

// group of Invarient
type Invariants []Invariant

// assertAll asserts the all invariants against application state
func (invs Invariants) assertAll(t *testing.T, app *baseapp.BaseApp,
	event string, displayLogs func()) {

	ctx := app.NewContext(false, abci.Header{Height: app.LastBlockHeight() + 1})

	for i := 0; i < len(invs); i++ {
		if err := invs[i](ctx); err != nil {
			fmt.Printf("Invariants broken after %s\n%s\n", event, err.Error())
			displayLogs()
			t.Fatal()
		}
	}
}

// GetAllAccounts returns all accounts in the accountKeeper.
func GetAllAccounts(mapper auth.AccountKeeper, ctx sdk.Context) []auth.Account {
	accounts := []auth.Account{}
	appendAccount := func(acc auth.Account) (stop bool) {
		accounts = append(accounts, acc)
		return false
	}
	mapper.IterateAccounts(ctx, appendAccount)
	return accounts
}
