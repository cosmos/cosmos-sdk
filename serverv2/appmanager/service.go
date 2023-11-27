package appmanager

type AppManager interface {
	Init() error
	DeliverBlock() error
}

/*
Things app manager needs to do:

Transaction:
- txdecoder
	- the transaction is already decoded in consensus there should not be a need to decode it again here
	- we should register the interface registstry to the txCodec
- txvalidator
	- needs to register the antehandlers

Execution: (DeliverBlock)
- execution of a transaction
	- Preblock call
	- BeginBlock call
		- PremessageHook
	- DeliverTx call
		- PostmessageHook
	- EndBlock call
- ability to register hooks
- ability to register messages and queries

Genesis:
- read genesis
- execute genesis txs

States:
- ExecuteTx
- SimulateTx

Queries:
- Query Router points to modules

Messages:
- Message Router points to modules

Config:
- QueryGasLimit
- HaltTime
- HaltBlock

Recovery:
- Panic Recovery for the app manager


*/
