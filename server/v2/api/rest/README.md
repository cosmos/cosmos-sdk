# Cosmos SDK REST API

This document describes how to use a service that exposes endpoints based on Cosmos SDK Protobuf message types. Each endpoint responds with data in JSON format.

## General Description

The service allows querying the blockchain using any type of Protobuf message available in the Cosmos SDK application through HTTP `POST` requests. Each endpoint corresponds to a Cosmos SDK protocol message (`proto`), and responses are returned in JSON format.

## Example

### 1. `QueryBalanceRequest`

This endpoint allows querying the balance of an account given an address and a token denomination.

- **URL:** `localhost:8080/cosmos.bank.v2.QueryBalanceRequest`

- **Method:** `POST`

- **Headers:**

  - `Content-Type: application/json`

- **Body (JSON):**

  ```json
  {
      "address": "<ACCOUNT_ADDRESS>",
      "denom": "<TOKEN_DENOMINATION>"
  }
  ```

  - `address`: Account address on the Cosmos network.
  - `denom`: Token denomination (e.g., `stake`).

- **Request Example:**

  ```
  POST localhost:8080/cosmos.bank.v2.QueryBalanceRequest
  Content-Type: application/json

  {
      "address": "cosmos16tms8tax3ha9exdu7x3maxrvall07yum3rdcu0",
      "denom": "stake"
  }
  ```

- **Response Example (JSON):**

  ```json
  {
      "balance": {
          "denom": "stake",
          "amount": "1000000"
      }
  }
  ```

  The response shows the balance of the specified token for the given account.

## Using Tools

### 1. Using `curl`

To make a request using `curl`, you can run the following command:

```bash
curl -X POST localhost:8080/cosmos.bank.v2.QueryBalanceRequest \
  -H "Content-Type: application/json" \
  -d '{
    "address": "cosmos16tms8tax3ha9exdu7x3maxrvall07yum3rdcu0",
    "denom": "stake"
  }'
```