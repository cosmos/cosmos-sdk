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

// assertAllInvariants asserts a list of provided invariants against
// application state
func assertAllInvariants(t *testing.T, app *baseapp.BaseApp,
	invariants []Invariant, where string, displayLogs func()) {

	for i := 0; i < len(invariants); i++ {
		err := invariants[i](app)
		if err != nil {
			fmt.Printf("Invariants broken after %s\n", where)
			fmt.Println(err.Error())
			displayLogs()
			t.Fatal()
		}
	}
}
