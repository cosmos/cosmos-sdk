# Group Module Architecture

## Overview

The Group module provides on-chain multisig and governance capabilities through groups, group policies, and proposals.

## Key Concepts

### Group
An aggregation of accounts with associated weights. Has an administrator who can add, remove, and update members.

### Group Policy
An account associated with a group and a decision policy. Can execute messages when proposals pass.

### Decision Policy
Defines how proposals are voted on and when they pass. Built-in policies: Threshold and Percentage.

### Proposal
A set of messages that will be executed if the proposal passes. Includes metadata, proposers, and voting windows.

## Module Dependencies

- **x/auth**: Account management, address codec
- **x/bank**: Token transfers for group policy execution

## Storage

Uses ORM (Object-Relational Mapping) with tables for:
- GroupInfo
- GroupMember
- GroupPolicyInfo
- Proposal
- Vote

## Configuration

- `MaxExecutionPeriod`: Max duration after voting ends to execute
- `MaxMetadataLen`: Max length for metadata fields (default: 255)
