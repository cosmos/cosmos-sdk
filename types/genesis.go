package types

import "encoding/json"

// function variable used to initialize application state at genesis
type InitStater func(ctx Context, state json.RawMessage) Error
