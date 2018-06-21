package config

//_____________________________________________________________________

// Configuration structure for command functions that share configuration.
// For example: init, init gen-tx and testnet commands need similar input and run the same code

// Storage for init gen-tx command input parameters
type GenTx struct {
	Name      string
	CliRoot   string
	Overwrite bool
	IP        string
}
