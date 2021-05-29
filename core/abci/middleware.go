package abci

type Middleware func(Handler) Handler
