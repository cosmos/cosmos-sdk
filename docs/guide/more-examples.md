
## Mintcoin


You just read about the amazing [plugin system](https://github.com/tendermint/basecoin/blob/develop/Plugins.md), and want to use it to print your own money.  Me too!  Let's get started with a simple plugin extension to basecoin, called [mintcoin](./mintcoin/README.md). This plugin lets you register one or more accounts as "central bankers", who can unilaterally issue more currency into the system.  It also serves as a simple test-bed to see how one can not just build a plugin, but also take advantage of existing codebases to provide a simple cli to use it.

## Financial Instruments

Sure, printing money and sending it is nice, but sometimes I don't fully trust the guy at the other end. Maybe we could add an escrow service? Or how about options for currency trading, since we support multiple currencies? No problem, this is also just a plugin away.  Checkout our [trader application](./trader).

**Running code, still WIP**

## IBC

Now, let's hook up your personal crypto-currency with the wide world of other currencies, in a distributed, proof-of-stake based exchange.  Hard, you say?  Well half the work is already done for you with the [IBC, InterBlockchain Communication, plugin](./ibc.md).  Now, we just need to get cosmos up and running and time to go and trade.

