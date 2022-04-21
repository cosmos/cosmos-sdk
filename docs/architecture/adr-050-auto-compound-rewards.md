# ADR 050: Auto-Compound Rewards

## Changelog

* April 17, 2022: Initial Draft

## Status

Draft (not implemented)

## Abstract

This ADR describes a modification of the `x/distribution` module's functionality
to allow users to request the ability to auto-compound their rewards to their
delegated validators on-chain.

## Context

As of SDK version v0.45.x, the `x/distribution` module defines a mechanism in
which delegators receive rewards by delegating voting power to validators in the
form of a native staking token. The reward distribution itself happens in a lazy
fashion and is defined by the [F1 specification](https://drops.dagstuhl.de/opus/volltexte/2020/11974/pdf/OASIcs-Tokenomics-2019-10.pdf).
In other words, delegators accumulate "unrealized" rewards having to explicitly
execute message(s) on-chain in order to withdraw said rewards. This provides the
ability for the chain to not have to explicitly distribute rewards to delegators
on a block-by-block basis which would otherwise make the network crawl to halt
as the number of delegations increases over time.

However, it has been shown that there is a strong desire to auto-compound
distribution rewards. As a result, there has been a proliferation of scripts, tooling
and clients off-chain to facilitate such a mechanism. However, these methods are
ad-hoc, often provide a cumbersome UX, and do not scale well to multiple networks.
