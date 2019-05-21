# Anatomy of an SDK Application

## Pre-requisite reading

- [High-level overview of an SDK application architecture](../intro/sdk-app-architecture.md)
- [Cosmos SDK design overview](../intro/sdk-design.md)

## Synopsis

This document describes the core parts of a Cosmos SDK application. 

- [Core Application File](#core-application-file)
- [Modules](#modules)
- [Client](#client)
- [Intefaces](#interfaces)
- [Dependencies and Makefile](#dependencies-and-makefile)

The core parts above will generally translate to the following file tree in the application directory:

```
./application
├── app.go
├── x/
├── cmd
│   ├── nsd
│   └── nscli
├── Gopkg.toml
└── Makefile
``` 

## Core Application File

In general, the core of the state-machine is defined in a file called `app.go`. 

## Modules (`./x/`)

## Client (Daemon)

## Interfaces

## Dependencies and Makefile 