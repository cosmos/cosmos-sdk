package types

// State implements a commitment state as defined in https://github.com/cosmos/ics/tree/master/spec/ics-023-vector-commitments#commitment-state.
// It represents full state of the commitment, which will be stored by the manager.
type State struct {
	vectorCommitment map[string][]byte
}

// NewState creates a new commitment State instance
func NewState(vectorCommitment map[string][]byte) State {
	return State{
		vectorCommitment: vectorCommitment,
	}
}

// Set adds a mapping from a path -> value to the vector commitment.
func (s *State) Set(path string, commitment []byte) {
	s.vectorCommitment[path] = commitment
}

// Remove removes a commitment under a specific path.
func (s *State) Remove(path string) {
	delete(s.vectorCommitment, path)
}
