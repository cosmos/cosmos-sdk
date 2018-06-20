# Getting Started

This document describes how to instantiate the LCD and connect to it. It provides examples and 
sample code.


## Configuration

To start a rest server, we need to specify the following parameters:

| Parameter | Type   | Default | Required | Description                          |
| --------- | ------ | ------- | -------- | ------------------------------------ |
| Chain-id  | string | null    | true     | chain id of the full node to connect |
| node      | URL | null    | true     | address of the full node to connect  |
| laddr      | URL | null    | true     | address to run the rest server on  |
| trust-store      | DIRECTORY | null    | true     | directory for save checkpoints and validator sets |