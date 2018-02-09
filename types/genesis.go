package types

import "encoding/json"

// function variable used to initialize application state at genesis
type InitStater func(ctxCheckTx, ctxDeliverTx Context, state json.RawMessage) Error
