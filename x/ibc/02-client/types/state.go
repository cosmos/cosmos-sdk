package types

// State is a type that represents the state of a client.
// Any actor holding the Stage can access on and modify that client information.
type State struct {
	// Client ID
	ID string `json:"id" yaml:"id"`
	// Boolean that states if the client is frozen when a misbehaviour proof is
	// submitted in the event of an equivocation.
	Frozen bool `json:"frozen" yaml:"frozen"`
}

// NewClientState creates a new ClientState instance
func NewClientState(id string) State {
	return State{
		ID:     id,
		Frozen: false,
	}
}
