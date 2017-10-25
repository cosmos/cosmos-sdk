/*
Package app contains data structures that provide basic
data storage functionality and act as a bridge between the abci
interface and the internal sdk representations.

StoreApp handles creating a datastore or loading an existing one
from disk, provides helpers to use in the transaction workflow
(check/deliver/commit), and provides bindings to the ABCI interface
for functionality such as handshaking with tendermint on restart,
querying the data store, and handling begin/end block and commit messages.
It does not handle CheckTx or DeliverTx, or have any logic for modifying
the state, and is quite generic if you don't wish to use the standard Handlers.

BaseApp embeds StoreApp and extends it for the standard sdk usecase, where
we dispatch all CheckTx/DeliverTx messages to a handler (which may contain
decorators and a router to multiple modules), and supports a Ticker which
is called every BeginBlock.
*/
package app
