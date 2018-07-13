# FAQ

## Overview

### How do I get Atoms?

If you participated in the fundraiser, you can check your suggested atom balance at [fundraiser.cosmos.network](https://fundraiser.cosmos.network).
If not, you must wait until the [Cosmos Network launches](/roadmap) and Atoms are traded on exchanges.

### Are Atoms listed on exchanges?

No. The Cosmos Network mainnet has not yet launched, which means Atoms are _not_ on exchanges. $CMOS and $ATOM tokens are _not_ Cosmos Network native tokens.

### How do I participate in the fundraiser?

The [fundraiser](https://fundraiser.cosmos.network) is closed. The Interchain Foundation raised funds from private individuals and has hosted a public fundraising event on which ended on April 6, 2017. Both $ETH and $BTC were accepted in the fundraiser. The security of the fundraising process has been vetted extremely carefully.

### What is the initial allocation of Atoms?

As a public, decentralized network, the allocation of Atoms is decided by those who run the software for the Cosmos Hub. To faciliate a decision, we are creating a Swiss non-profit, the [Interchain Foundation](https://interchain.io), which is responsible for co-ordinating fundraising and allocating funds to get the network off the ground. The foundation will suggest a allocation of Atoms according to the results of the fundraiser. Users will ultimately decide the distribution for themselves when they run the software.

The Interchain Foundation will suggest that 5% of the Atoms go to its initial donors, 10% go to the Interchain Foundation, 10% go to the company developing most of the software, and the remaining 75% to be distributed according to the results of the private and public fundraisers.

### What is the team developing the Cosmos Network?

The Cosmos Network is the first project being funded by the Interchain Foundation. Its development is led primarily by the [Tendermint team](/about/team).

### What's the difference between Tendermint, the Cosmos Network, and the Cosmos Hub?

- [Tendermint](https://tendermint.com) is a general purpose blockchain engine that uses a Byzantine-fault tolerant consensus protocol and allows applications to be written in any programming language.
- The Cosmos Network is a heterogenous network of Proof-of-Stake blockchains that can interoperate with one-another.
- The Cosmos Hub is the first Proof-of-Stake blockchain to be launched by the Cosmos Network; it uses Tendermint consensus, contains a built in governance protocol, and serves as co-ordinater for interoperability between other blockchains.
- Atoms: The native cryptocurrency on the Cosmos Hub. Atoms are necessary for participating in the consensus protocol and transacting on the network.

### When will the Cosmos Network launch?

Please check [our roadmap](https://cosmos.network/roadmap).

### What is the utility of Atoms?

Public, decentralized networks require high levels of security and spam-prevention that are best achieved by economic means: participants in the consensus must incur some economic cost, and all transactions processed by the network must pay a fee. Since we want to use Proof-of-Stake validators instead of Proof-of-Work miners, we require validators of the Cosmos Hub to make a large security deposit in Atoms - if they misbehave, their Atoms are revoked by the protocol!

The more Atoms in security deposits, the more stake on the line; the more skin-in-the-game; the greater the economic security. In this sense, the Atoms act like virtual miners.

To achieve spam-prevention, all transactions on the Cosmos Hub must pay a fee in Atoms. The fee may be proportional to the amount of computation required by the transaction, similar to Ethereum's concept of "gas". The fees are collected by the validators and distributed proportionately to the Atoms held in security deposits.

## Interoperability

### What's an IBC packet?

[IBC packets](https://blog.cosmos.network/developer-deep-dive-cosmos-ibc-5855aaf183fe) are packets of data that one blockchain wishes to send to another blockchain. But instead of literally sending a packet of bytes via the TCP/IP or UDP/IP protocol (which is designed for singular, physical, machines), IBC packets require cryptographic proof-of-existence. Since no single node or validator has the authority to speak on behalf of the entire blockchain, and, since we don't want to rely on the integrity of the IP internet infrastructure, instead we rely on a cryptographic proof of a blockchain hash commit (+2/3 of signatures for that blockchain hash) along with a Merkle-proof from the aforementioned blockhash to a packet in the blockchain's "application state", which proves that the blockchain validators agreed to publish this packet of information. So, anyone who sees an IBC packet (regardless of the source of this data) can verify its integrity.

### How does one exchange currencies in this system?

For tokens outside the Cosmos system, they can only be introduced via pegged
derivatives. Read about interoperating with existing blockchains here: [Peggy](https://blog.cosmos.network/the-internet-of-blockchains-how-cosmos-does-interoperability-starting-with-the-ethereum-peg-zone-8744d4d2bc3f).

```
           _ peg smart contract
          /
  [  Ethereum  ] <--> [ EtherCosmos Peg Zone ] <-IBC-> [  Cosmos Hub  ] <-IBC-> (Bitcoin) [ PoW/Casper ]
                      [      Tendermint      ]         [  Tendermint  ] <-IBC-> (exchange)
```

### How does Cosmos manage governance?

In Cosmos, the stakeholders are well defined, as is the prior social contract. Ethereum had a hard time with the fork because they had to ask the ether holders as well as the miners, but the ether holders had no prior social contract or obligation to partake in governance, so no quorum could be reached in time. Asking the miners is necessary to ensure that the hard-fork will have support, but after a while they tend to simply follow the money and incentives.

Cosmos is different because instead of anonymous miners we have social contract bound validators and delegators who have stake, and, they have the obligation to partake in governance.

## Validators

### What is the maximum number of validators in Cosmos? What about nodes?

We will start with 100 validators. Anyone else can be a node. To start, the validators will be the same across all shards - they will run the shards concurrently. Over time, these restrictions will be loosened. Misbehaviour in the consensus on any shard will result in security deposits being revoked.

### What will be the process for abandoning validators that misbehave?

If a validator misbehaves on its own by double-signing at the same height &amp; round, then the evidence is very short and simple -- it's just the two conflicting votes. This evidence can be included in the the Cosmos Hub as a Slash transaction, and the validator will immediately become inactive and slashed after the Slash transaction gets committed.

If there is a zone fork, either of the Cosmos Hub or any of the zones, the two conflicting commits also constitute evidence. This is a much more complicated data structure. It is guaranteed to slash at least 1/3 of the validators' atoms for that zone.

### What's the difference between a Delegator and a Validator?

A [validator](/staking/validators) has an active key involved in signing votes in the consensus protocol. A validator must also have some Atoms in a security deposit. Since there will only be a limitted number of validators, [other Atom holders can delegate](/staking/delegators) to the validators, thereby contributing to the economic security of the system by putting their funds on the line if the validator misbehaves. In return, they earn a share of the transaction fees and any inflationary rewards.

### Can delegators also be validators?

Delegators are never validators. If a validator wishes to delegate, they need to do so with their free and unbonded Atoms.

### How are validator voting powers determined and changed?

Validators are initially determined according to a public vote among Atom holders to be carried out before the launch of the Cosmos Hub. Atom holders delegate to the various candidates, and the top 100 candidates will be the initial validators. Once [the Hub launches](/roadmap), the vote will be a continuous process where users shuffle around their delegated Atoms, thereby changing the validator set.

Part of the purpose of the fundraiser is to distribute Atoms across a wide variety of individuals and organizations so that the validator set will be sufficiently decentralized for a robust network. In the event of attacks or mishaps, the blockchain may need to purge bad actors through socially co-ordinated hard-forks. The ability to account for misbehaviour and co-ordinate hardforks helps make the system antifragile.
