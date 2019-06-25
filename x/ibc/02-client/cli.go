package client

import ()

// CLIObject stores the key for each object fields
type CLIObject struct {
	ID             string
	ConsensusState []byte
	Frozen         []byte
}

func (object Object) CLI() CLIObject {
	return CLIObject{
		ID:             object.id,
		ConsensusState: object.consensusState.Key(),
		Frozen:         object.frozen.Key(),
	}
}
