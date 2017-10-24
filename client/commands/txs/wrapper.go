package txs

import (
	"github.com/spf13/pflag"
)

var (
	// Middleware must be set in main.go to defined the wrappers we should apply
	Middleware Wrapper
)

// Wrapper defines the information needed for each middleware package that
// wraps the data.  They should read all configuration out of bounds via viper.
type Wrapper interface {
	Wrap(interface{}) (interface{}, error)
	Register(*pflag.FlagSet)
}

// Wrappers combines a list of wrapper middlewares.
// The first one is the inner-most layer, eg. Fee, Nonce, Chain, Auth
type Wrappers []Wrapper

var _ Wrapper = Wrappers{}

// Wrap applies the wrappers to the passed in tx in order,
// aborting on the first error
func (ws Wrappers) Wrap(tx interface{}) (interface{}, error) {
	var err error
	for _, w := range ws {
		tx, err = w.Wrap(tx)
		if err != nil {
			break
		}
	}
	return tx, err
}

// Register adds any needed flags to the command
func (ws Wrappers) Register(fs *pflag.FlagSet) {
	for _, w := range ws {
		w.Register(fs)
	}
}
