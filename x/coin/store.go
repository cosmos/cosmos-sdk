package coin

import "github.com/tendermint/go-wire/data"

// TEMP

type Actor struct {
	ChainID string
	App     string
	Address data.Bytes
}

// Account - coin account structure
type Account struct {
	// Coins is how much is on the account
	Coins Coins `json:"coins"`
	// Credit is how much has been "fronted" to the account
	// (this is usually 0 except for trusted chains)
	Credit Coins `json:"credit"`
}
