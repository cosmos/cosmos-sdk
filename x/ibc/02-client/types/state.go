package types

// ClientState is a type that represents the state of a client.
// Any actor holding the Stage can access on and modify that client information.
type ClientState struct {
	// Client ID
	id string
	// Boolean that states if the client is frozen when a misbehaviour proof is
	// submitted in the event of an equivocation.
	Frozen bool `json:"frozen" yaml:"frozen"`
}

// NewClientState creates a new ClientState instance
func NewClientState(id string) ClientState {
	return ClientState{
		id:     id,
		Frozen: false,
	}
}

// ID returns the client identifier
func (cs ClientState) ID() string {
	return cs.id
}
