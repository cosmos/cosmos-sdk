# Plugin Examples

Now that we've seen how to use Basecoin, talked about the design, 
and looked at how to implement a simple plugin, let's take a look at some more interesting examples.

## Mintcoin

Basecoin does not provide any functionality for adding new tokens to the system.
The state is endowed with tokens by a `genesis.json` file which is read once when the system is first started.
From there, tokens can be sent to other accounts, even new accounts, but it's impossible to add more tokens to the system.
For this, we need a plugin.

The `mintcoin` plugin lets you register one or more accounts as "central bankers", 
who can unilaterally issue more currency into the system.  

## Financial Instruments

Sure, printing money and sending it is nice, but sometimes I don't fully trust the guy at the other end. Maybe we could add an escrow service? Or how about options for currency trading, since we support multiple currencies? No problem, this is also just a plugin away.  Checkout our [trader application](./trader).

**Running code, still WIP**

## IBC

Now, let's hook up your personal crypto-currency with the wide world of other currencies, in a distributed, proof-of-stake based exchange.  Hard, you say?  Well half the work is already done for you with the [IBC, InterBlockchain Communication, plugin](./ibc.md).  Now, we just need to get cosmos up and running and time to go and trade.

