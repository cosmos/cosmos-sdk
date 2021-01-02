<!--
order: 0
title: Nameservice Overview
parent:
  title: "nameservice"
-->

# `x/nameservice`

## Table of Contents

<!-- TOC -->

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
4. **[Params](04_params.md)**


## Abstract

`x/nameservice` is an implementation of a Cosmos SDK module,
that improves decentralization, security, censorship resistance, privacy, and speed of certain components of the Internet infrastructure such as Domain Name Service and identities.

The nameservice module helps you to build an application that lets users buy names and to set a value these names resolve to. Users can buy unused names, or sell/trade their name. The owner of a given name is the current highest bidder. Also, allows the owner of the name to delete a given name.

To build a nameservice application, you need to declare a few modules in `app.go` file: `auth`, `bank`, `staking`, `distribution`, `slashing` and `nameservice`. The first five already exist, but not the nameservice module . You have to build the nameservice module that defines the bulk of the state machine.