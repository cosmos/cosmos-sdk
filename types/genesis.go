package types

// function variable used to initialize application state at genesis
type InitStater func(stateJSON []byte) Error
