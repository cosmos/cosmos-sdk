# Implementing Account Abstraction

In order to implement account abstraction an account is required to register a handler which accepts a MsgAuthenticate.

Please refer to the type documentation to understand what each field in the UserOperation means.

// TODO: do not be lazy and add here.

In the SDK an externally owned account (EOA for short) can execute a state transition in two ways:
- through a bundler: the user can send a UserOperation to a bundler and then the bundler would send a tx that in turn executes the user operation.
- through a tx: in this case we call the AccountAbstraction implementation enshrined.

## Authentication

At the core of the authentication lives the UserOperation, and that is what should be used to authenticate a User.
Since 
