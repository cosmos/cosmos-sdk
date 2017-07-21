# Quark

Quarks are the fundamental building blocks of atoms, through which DNA, life,
and matter arise. Similarly this package is the core framework for constructing
the atom tokens which will power [The Cosmos Network](https://cosmos.network/).

The Quark framework affords you all the tools you need to rapidly develop
robust blockchains and blockchain applications which are interoperable with The
Cosmos Hub. Quark is an abstraction of [Tendermint](https://tendermint.com/)
which provides the core consensus engine for your blockchain. Beyond consensus,
Quark provides a blockchain development 'starter-pack' of common blockchain
modules while not enforcing their use thus giving maximum flexibility for
application customization (do you require fees, how do you want to log
messages, do you enable IBC, do you even have a cryptocurrency?)

Disclaimer, when power and flexibility meet, the result is also some level of
complexity and a learning curve.  Here is an introduction to the core concepts
embedded in Quark.

## Inspiration

The basic concept came from years of web development.  A number of patterns
have arisen in that realm of software which enable people to build remote
servers with APIs remarkably quickly and with high stability.  The
[ABCI](https://github.com/tendermint/abci) application interface is similar to
a web API (DeliverTx is like POST and Query is like GET and `SetOption` is like
the admin playing with the config file). Here are some patterns that might be
useful:

* MVC - separate data model (storage) from business logic (controllers)
* Routers - easily direct each request to the appropriate controller
* Middleware - a series of wrappers that provide global functionality (like
  authentication) to all controllers
* Modules (gems, package, ...) - developers can write a self-contained package
  with a given set of functionality, which can be imported and reused in other
  apps

Also, the idea of different tables/schemas in databases, so you can keep the
different modules safely separated and avoid any accidental (or malicious)
overwriting of data.

Not all of these can be pulled one-to-one in the blockchain world, but they do
provide inspiration to provide orthogonal pieces that can easily be combined
into various applications.

## Further reading

* [Glossary of the terms](glossary.md)
* [Standard modules](stdlib.md)
* Guide to building a module
* Demo of CLI tool
* IBC in detail
* Diagrams... Coming Soon!
