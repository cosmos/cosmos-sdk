

TODO: write a blog post ...

# Inspiration

The basic concept for this SDK comes from years of web development. A number of
patterns have arisen in that realm of software which enable people to build remote
servers with APIs remarkably quickly and with high stability. The
[ABCI](https://github.com/tendermint/abci) application interface is similar to
a web API (`DeliverTx` is like POST and `Query` is like GET while `SetOption` is like
the admin playing with the config file). Here are some patterns that might be
useful:

* MVC - separate data model (storage) from business logic (controllers)
* Routers - easily direct each request to the appropriate controller
* Middleware - a series of wrappers that provide global functionality (like
  authentication) to all controllers
* Modules (gems, package, etc) - developers can write a self-contained package
  with a given set of functionality, which can be imported and reused in other
  apps

Also at play is the concept of different tables/schemas in databases, thus you can
keep the different modules safely separated and avoid any accidental (or malicious)
overwriting of data.

Not all of these can be compare one-to-one in the blockchain world, but they do
provide inspiration for building orthogonal pieces that can easily be combined
into various applications.
