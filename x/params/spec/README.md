<!--
order: 0
title: Params Overview
parent:
  title: "params"
-->

# `params`

## Abstract

Package params provides a globally available parameter store.

There are two main types, Keeper and Subspace. Subspace is an isolated namespace for a
paramstore, where keys are prefixed by preconfigured spacename. Keeper has a
permission to access all existing spaces.

Subspace can be used by the individual keepers, which need a private parameter store
that the other keepers cannot modify. The params Keeper can be used to add a route to `x/gov` router in order to modify any parameter in case a proposal passes.

The following contents explains how to use params module for master and user modules.

## Contents

1. **[Keeper](01_keeper.md)**
2. **[Subspace](02_subspace.md)**
    - [Key](02_subspace.md#key)
    - [KeyTable](02_subspace.md#keytable)
    - [ParamSet](02_subspace.md#paramset)
