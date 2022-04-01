# Cosmos-SDK Roadmap 2022

The Cosmos-SDK has gone through many interations over the past year, from rewriting how governance works, to writing new modules we have seen this softward evolve. With 2022 already under way we are preparing to release 0.46 and move twoards new goals for the software. 

There has been work towards many of the 2022 goals already but many will still need to be ironed out. 

## Goals

- Performance
- Maintainability
- Comprehension

## Features

- Proto
	- use new proto generator
- ORM
	- module rewrite to include this
- Framework
	 - how user create an application
	 - Bech32 refactor
		- removing global
		- passing around address codec
		- baseapp++ (v2)`
- Dynamic cosmos (grpc)
- Keystone
	- replace keyring
- Storage
	- make storage more efficient
	- smt
	- SC & SS
- Sign mode Textual
	- 
- Vesting account refactor
	- single account many vesting schedules
	- pull out vesting into its own module



> Note:
	- small fast iteration
	- small releases, faster
