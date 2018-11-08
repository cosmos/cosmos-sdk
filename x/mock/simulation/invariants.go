package simulation

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

// An Invariant is a function which tests a particular invariant.
// If the invariant has been broken, it should return an error
// containing a descriptive message about what happened.
// The simulator will then halt and print the logs.
type Invariant func(app *baseapp.BaseApp) error

// group of Invarient
type Invariants []Invariant

// assertAll asserts the all invariants against application state
func (invs Invariants) assertAll(t *testing.T, app *baseapp.BaseApp,
	event string, displayLogs func()) {

	for i := 0; i < len(invs); i++ {
		if err := invs[i](app); err != nil {
			fmt.Printf("Invariants broken after %s\n%s\n", event, err.Error())
			displayLogs()
			t.Fatal()
		}
	}
}
