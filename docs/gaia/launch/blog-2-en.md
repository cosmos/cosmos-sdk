# The 3 Phases of the Cosmos Hub Mainnet
## Post-Mainnet Development Roadmap & Expectations for Users

The launch of the Cosmos Hub mainnet is expected to happen in phases. Here we outline what to expect in each phase.

# 🚨Phase I: Network Gains Stability 🚨

In the first phase, the network is likely to be unstable; it may experience halts or other forms of failure requiring intervention and coordination among Cosmos Hub validators and full node operators to deploy a fix. This type of failure is not unexpected while the network gains stability. 

## State Reversions and Mainnet launch

One of the core ideologies around blockchains is immutability. This is the idea that we don't go 
back and edit past state transitions. While this notion of immutability is implemented directly via consensus protocols in the software, it is ultimately upheld by social contract among participants.

That said, the technology underlying the Cosmos Hub was intentionally developed to enable low-friction forks and rollbacks. We’ve seen the community practice these techniques numerous times on the test networks. It’s likely they will need to be used on a mainnet as well. Ultimately, they are a countervailing force to the risk of cartel takeover.

Reverting state is often seen as highly grievous, as it compromises the network’s economic finality. Hence it should only be used in extreme conditions, as witnessed in the case of Ethereum with the DAO Hard Fork. That said, in the early days of the Cosmos Hub network, transfers will not be active, and hence the severity of state reversions will be reduced, as state transitions will be much less “economically final”. If necessary in case of bugs, the state can be exported from a past height and the network restarted, as practiced on the testnets.

Once governance chooses to enable transfers, the importance of economic finality must be respected by the network.

To summarize, if there are errors or vulnerabilities in the Cosmos Hub in the days before transfers are enabled, users should expect arbitrary state rollbacks even to genesis.

Once transfers are enabled, state rollbacks will be much more difficult to justify. 

**What this means for developers:** The Cosmos mainnet launch is the first phase in which fundraiser participants will be working together to operate the software. As a decentralized application developer, you are likely a user of either the [Cosmos-SDK framework](https://cosmos.network/docs/) or [Tendermint Core](https://tendermint.com/docs/). The progress of your Cosmos-SDK or Tendermint-based application should be independent of the Cosmos Hub roadmap. However, if your project requires the use of [Inter-Blockchain Communication][blog post], you must wait until Phase III, or participate in the IBC testnets that will begin shortly.

**What this means for users:** In this phase, we strongly recommend that you do not arrange to trade Atoms (eg. by legal contract as they will not be transferable yet) as there is the risk of state being reverted.

You can, however, safely delegate Atoms to validators in this phase by following the CLI guideline and video tutorial linked below. Of course, in the event of a state reversion, any earned fees and inflation may be lost. Note that only `gaiacli` should be used for making transactions. Voyager, the GUI for interacting with the Cosmos Hub, is currently in alpha and undergoing development. A separate announcement will be made once Voyager is safer for use.

CLI Guide 🔗:  [github.com/cosmos/cosmos-sdk/…/delegator-guide-cli.md](https://github.com/cosmos/cosmos-sdk/blob/develop/docs/gaia/delegator-guide-cli.md) 

**Watch CLI delegation tutorial:** [Cosmos YouTube](https://www.youtube.com/watch?v=ydZw6o6Mzy0)

# Phase II: Transfers Enabled

**Summary:** Once mainnet is deemed sufficiently stable, bonded Atom holders will vote to decide whether or not Atom transfers should be enabled. This procedure will happen through on-chain governance.

The best way to check on the status of governance proposals is to view them through Cosmos explorers. A list of explorers can be found on the launch page: [cosmos.network/launch](https://cosmos.network/launch). 

**What this means for users:** If the proposal is accepted and transfers are enabled, then it becomes possible to transfer Atoms.

# Phase III: IBC Enabled

**Summary:** In Phase III, the [IBC protocol][ibc] is released and Atom holders vote via on-chain governance on whether or not to enable it as part of the core module library within the Cosmos-SDK. 

**What this means for developers:** Application-specific blockchains that are built using the Cosmos-SDK or Tendermint BFT will be able to connect to the Hub and interoperate/compose with all of the other blockchains that are connected to it.

**What this means for users:** You will be able to transfer various tokens and NFTs directly from one IBC-connected chain to another IBC-connected chain without going through a centralized 
third-party platform.

# Housekeeping for Validators: Submitting a `gentx` for mainnet

1. You should have generated and secured the validator consensus key you are going to be validating under during mainnet.
2. Be prepared to sign a transaction under an address in the genesis file either from the fundraiser or Game of Stakes depending on where you received your ATOM allocation.
3. We will begin collecting Gentxs for mainnet once the recommended genesis allocations are published.

# In Closing
The Cosmos mission is to build bridges to connect all blockchains—to build an Internet of Blockchains. Clearly, we have a long road of development ahead of us. And after mainnet, the real work to a world of deeply integrated token economies is still ahead of us. But as John Fitzgerald Kennedy once said in the face of adversity:

*“We choose to go to the moon...not because they are easy, but because they are hard….”*

To the Moon 🚀


[blog post]: [https://blog.cosmos.network/developer-deep-dive-cosmos-ibc-5855aaf183fe]
[ibc]: [https://github.com/cosmos/cosmos-sdk/blob/develop/docs/spec/ibc/overview.md]
