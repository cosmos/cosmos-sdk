SDK Overview
==============

The SDK middleware design optimizes flexibility and security. The
framework is designed around a modular execution stack which allows
applications to mix and match modular elements as desired. Along side,
all modules are permissioned and sandboxed to isolate modules for
greater application security.

Framework Overview
------------------

Transactions
~~~~~~~~~~~~

Each transaction passes through the middleware stack which can be
defined uniquely by each application. From the multiple layers of
transaction, each middleware may strip off one level, like an onion. As
such, the transaction must be constructed to mirror the execution stack,
and each middleware module should allow an arbitrary transaction to be
embedded for the next layer in the stack.

Execution Stack
~~~~~~~~~~~~~~~

Middleware components allow for code reusability and integrability. A
standard set of middleware are provided and can be mix-and-matched with
custom middleware. Some of the `standard library <./stdlib.html>`__
middlewares provided in this package include: - Logging - Recovery -
Signatures - Chain - Nonce - Fees - Roles -
Inter-Blockchain-Communication (IBC)

As a part of stack execution the state space provided to each middleware
is isolated in the ``Data Store`` below. When
executing the stack, state-recovery checkpoints can be assigned for
stack execution of ``CheckTx`` or ``DeliverTx``. This means, that all
state changes will be reverted to the checkpoint state on failure when
either being run as a part of ``CheckTx`` or ``DeliverTx``. Example
usage of the checkpoints is when we may want to deduct a fee even if the
end business logic fails; under this situation we would add the
``DeliverTx`` checkpoint after the fee middleware but before the
business logic. This diagram displays a typical process flow through an
execution stack.



Dispatcher
~~~~~~~~~~

The dispatcher handler aims to allow for reusable business logic. As a
transaction is passed to the end handler, the dispatcher routes the
logic to the correct module. To use the dispatcher tool, all transaction
types must first be registered with the dispatcher. Once registered the
middleware stack or any other handler can call the dispatcher to execute
a transaction. Similarly to the execution stack, when executing a
transaction the dispatcher isolates the state space available to the
designated module (see ``Data Store`` below).

Security Overview
-----------------

Permission
~~~~~~~~~~

Each application is run in a sandbox to isolate security risks. When
interfacing between applications, if one of those applications is
compromised the entire network should still be secure. This is achieved
through actor permissioning whereby each chain, account, or application
can provided a designated permission for the transaction context to
perform a specific action.

Context is passed through the middleware and dispatcher, allowing one to
add permissions on this app-space, and check current permissions.

Data Store
~~~~~~~~~~

The entire merkle tree can access all data. When we call a module (or
middleware), we give them access to a subtree corresponding to their
app. This is achieved through the use of unique prefix assigned to each
module. From the module's perspective it is no different, the module
need-not have regard for the prefix as it is assigned outside of the
modules scope. For example, if a module named ``foo`` wanted to write to
the store it could save records under the key ``bar``, however, the
dispatcher would register that record in the persistent state under
``foo/bar``. Next time the ``foo`` app was called that record would be
accessible to it under the assigned key ``bar``. This effectively makes
app prefixing invisible to each module while preventing each module from
affecting each other module. Under this model no two registered modules
are permitted to have the same namespace.
