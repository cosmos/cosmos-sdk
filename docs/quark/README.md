# Quark

Quarks are the building blocks of atoms. And in a similar vein, this
package is a framework for building the cosmos. It gives you all the tools
you need to quickly build up powerful abci applications to run on tendermint,
while also providing maximum flexibility to customize aspects of your
application (do you require fees, how do you want to log messages, do you
enable IBC, do you even have a cryptocurrency?)

However, when power and flexibility meet, the result is also some level of
complexity and a learning curve.  Here is an introduction to the core concepts
embedded in quarks, so you can apply them properly.

## Inspiration

The basic concept came from years of web development.  After decades of web
development, a number of patterns have arisen that enabled people to build
remote servers with APIs remarkably quickly and with high stability.
I think the ABCI app interface is similar to a web api (DeliverTx is like POST
and Query is like GET and SetOption is like the admin playing with the config
file). Here are some patterns that might be useful:

* MVC - separate data model (storage) from business logic (controllers)
* Routers - easily direct each request to the appropriate controller
* Middleware - a series of wrappers that provide global functionality (like
  authentication) to all controllers
* Modules (gems, package, ...) - people can write a self-contained package
  with a given set of functionality, which can be imported and reused in
  other apps

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
* Demo of cli tool
* IBC in detail
* Diagrams!!!
