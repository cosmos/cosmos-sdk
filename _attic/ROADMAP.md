# Roadmap for future basecoin development

Warning: there are current plans, they may change based on other developments, needs.  The further in the future, the less clear, like all plans.

## 0.6.x - Testnet and Light Client (late June 2017)

The current release cycle is making sure the server is usable for deploying testnets (easy config, safe restarts, moving nodes). Also that we have a useable light client that does full cryptographic prooofs without syncing the entire block headers.  See the [changelog](CHANGELOG.md).

Patch release here involve improving the usability of the cli tools, adding subcommands, more flags, helper functions, shell integrations, etc.  Please add issues if you find the client tool difficult to use, or deployment troublesome.

## 0.7.x - Towards a modular framework (late July 2017)

**Breaking changes**

* Renaming likely: this release may well lead to a [renaming of the repository](https://github.com/tendermint/basecoin/issues/119) to emphasize that it is a generalized framework.  `basecoin` and `basecli` executables will remain with generally unchanged usage.
* This will also provide a tx structure that is very different than the current one, and a non-trivial upgrade of running chains.

The next release cycle involves a big upgrade to the core, especially how one can write modules (aka plugins) as well as configure a basecoin-based executable.  The main goal is to leave us with basecoin as a single executable with a similar API, but create a new module/middleware system with a number of standard modules provided (and easy addition of third party modules), so developers can quickly mix-and-match pieces and add custom business logic for there chain.

The main goal here is to migrate from a basecoin with plugins for extra enhancements, to a proper app development framework, of which basecoin is one example app that can quickly be built.

Some ideas:

* Flexible fee/gas system (good for both public and private blockchains)
* Flexible authentication systems (with multi-sig support)
* Basic role permission system
* Abstract IBC to support other transactions from various modules (not just sendtx)

This will be done in conjunction with some sample apps also building on this framework, where other logic is interesting and money transfers is not the central goal, like [trackomatron](https://github.com/tendermint/trackomatron)

## Next steps

The following are three planned steps, the order of which may change.  At least one or two of these will most likely occur before any other developments. Clearly, any other feature that are urgent for cosmos can jump the list in priority, but all these pieces are part of the cosmos roadmap, especially the first two.

### 0.8.x??? - Local client API for UI

Beyond the CLI, we want to add more interfaces for easily building a UI on top of the basecoin client.  One clear example is a local REST API, so you can easily integrate with an electron app, or a chrome app, just as if you wrote a normal Single-Page Application, but connecting to a local proxy to do full crypto-graphic proofs.

Another **possible** development is providing an SDK, which we can compile with [go-mobile](https://github.com/golang/go/wiki/Mobile) for both Android and iOS to support secure mobile applications. Progress on this front is contingent on participation of an experienced mobile developer.

Further, when the planned enhancements to real-time events happen in tendermint core, we should expose that via a simple subscriber/listener model in these local APIs.

### 0.9.x??? - Proof of Stake and Voting Modules

We should integrate developments on a [proof-of-stake module](https://github.com/tendermint/basecoin-stake) (currently a work-in-progress) and basic voting modules (currently planned) into properly supported for modules.  These would provide the basis for dynamic validator set changes with bondign periods, and the basis for making governance decisions (eg. voting to change the block reward).

At this point we would have to give full support to these plugins, and third-party devs can build on them to add more complex delegation or governance logic.

### 0.10.x??? - Database enhancements

Depending on developments with merkleeyes, we would like to increase the expressiveness of the storage layer while maintaining provability of all queries.  We would also add a number of new primatives to the key-value store, to allow some general data-structures.

Also, full support for historical queries and performance optimizations of the storage later.  But this all depends on supporting developments of another repo, so timing here is unclear.  Here are some example ideas:

Merkle proofs:

* **Proof of key-value**: only current proof
* **Proof of missing key**: prove there is no data for that key
* **Proof of range**: one proof for all key-values in a range of keys
* **Proof of highest/lowest in range**: just get one key, for example, prove validator hash with highest height <= H

Data structures:

* **Queues**: provable push-pop operations, split over multiple keys, so it can scale to 1000s of entries without deserializing them all every time.
* **Priority Queues**: as above, but ordered by priority instead of FIFO.
* **Secondary Indexes**: add support for secondary indexes with proofs. So, I can not only prove my balance, but for example, the list of all accouns with a balance of > 1000000 atoms. These indexes would have to be created by the application and stored extra in the database, but if you have a common query that you want proofs/trust, it can be very useful.
