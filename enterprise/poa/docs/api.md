# POA Module API Documentation

## Overview

The Proof of Authority (POA) module provides a governance mechanism for managing validators in a Cosmos SDK blockchain. Unlike traditional Proof of Stake, POA allows a designated admin to control validator set membership and voting power distribution.

**Package:** `cosmos.poa.v1`
**Go Import:** `github.com/cosmos/cosmos-sdk/enterprise/poa/types`

---

## Core Concepts

- **Admin Control:** A single admin address has exclusive authority to manage validators and module parameters
- **Validator Management:** Create validators, update voting power, and manage the active validator set
- **Fee Distribution:** Validators accumulate fees that can be withdrawn by their operators
- **Dynamic Updates:** Changes to the validator set are applied without stopping the chain

---

## Data Types

### Validator

Represents a validator in the POA system.

```protobuf
message Validator {
  google.protobuf.Any pub_key = 1;
  int64 power = 2;
  ValidatorMetadata metadata = 3;
  repeated cosmos.base.v1beta1.DecCoin allocated_fees = 4;
}
```

**Fields:**
- `pub_key` (Any): The validator's consensus public key (typically `/cosmos.crypto.ed25519.PubKey`)
- `power` (int64): Voting power for this validator (use `0` to remove a validator)
- `metadata` (ValidatorMetadata): Additional validator information
- `allocated_fees` (DecCoin[]): Accumulated fees allocated to this validator

**Example:**
```json
{
  "pub_key": {
    "@type": "/cosmos.crypto.ed25519.PubKey",
    "key": "YUzyiqZzKN8BmLbl75gdXfbxQ2QtSYpPSwA85bZ3xuE="
  },
  "power": "10000",
  "metadata": {
    "moniker": "validator-1",
    "description": "First validator node",
    "operator_address": "cosmos1x0mm8rws8lm46xay3zyyznzr6lvu5um3kht0x7"
  },
  "allocated_fees": []
}
```

### ValidatorMetadata

Metadata information about a validator.

```protobuf
message ValidatorMetadata {
  string moniker = 3;
  string description = 4;
  string operator_address = 5;
}
```

**Fields:**
- `moniker` (string): Human-readable name for the validator
- `description` (string): Optional description of the validator
- `operator_address` (string): Cosmos SDK address that operates this validator

### Params

Module parameters.

```protobuf
message Params {
  string admin = 1;
}
```

**Fields:**
- `admin` (string): Cosmos SDK address with administrative privileges

### ValidatorFees

Represents fee allocations for a validator operator.

```protobuf
message ValidatorFees {
  repeated cosmos.base.v1beta1.DecCoin fees = 1;
}
```

**Fields:**
- `fees` (DecCoin[]): List of coins representing allocated fees

---

## Query API

The Query service provides read-only access to POA module state.

### Params

Get module parameters.

**gRPC:** `cosmos.poa.v1.Query/Params`
**REST:** `GET /cosmos/poa/v1/params`

**Request:**
```protobuf
message QueryParamsRequest {}
```

**Response:**
```protobuf
message QueryParamsResponse {
  Params params = 1;
}
```

**CLI:**
```bash
simd q poa params
```

**Example Response:**
```json
{
  "params": {
    "admin": "cosmos1x0mm8rws8lm46xay3zyyznzr6lvu5um3kht0x7"
  }
}
```

---

### Validator

Query a single validator by address.

**gRPC:** `cosmos.poa.v1.Query/Validator`
**REST:** `GET /cosmos/poa/v1/validator/{address}`

**Request:**
```protobuf
message QueryValidatorRequest {
  string address = 1;  // Consensus or operator address
}
```

**Response:**
```protobuf
message QueryValidatorResponse {
  Validator validator = 1;
}
```

**CLI:**
```bash
simd q poa validator <address>
```

**Notes:**
- `address` can be either a consensus address or operator address

---

### Validators

List all validators in the system.

**gRPC:** `cosmos.poa.v1.Query/Validators`
**REST:** `GET /cosmos/poa/v1/validators`

**Request:**
```protobuf
message QueryValidatorsRequest {
  cosmos.base.query.v1beta1.PageRequest pagination = 2;
}
```

**Response:**
```protobuf
message QueryValidatorsResponse {
  repeated Validator validators = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
```

**CLI:**
```bash
simd q poa validators
```

**Notes:**
- Results are always returned in descending order by voting power
- Supports pagination for large validator sets

**Example Response:**
```json
{
  "validators": [
    {
      "pub_key": {
        "@type": "/cosmos.crypto.ed25519.PubKey",
        "key": "YUzyiqZzKN8BmLbl75gdXfbxQ2QtSYpPSwA85bZ3xuE="
      },
      "power": "10000",
      "metadata": {
        "moniker": "validator-1",
        "operator_address": "cosmos1..."
      }
    }
  ]
}
```

---

### WithdrawableFees

Query fees available for withdrawal by a validator operator.

**gRPC:** `cosmos.poa.v1.Query/WithdrawableFees`
**REST:** `GET /cosmos/poa/v1/allocated_fees/{operator_address}`

**Request:**
```protobuf
message QueryWithdrawableFeesRequest {
  string operator_address = 1;
}
```

**Response:**
```protobuf
message QueryWithdrawableFeesResponse {
  ValidatorFees fees = 1;
}
```

**CLI:**
```bash
simd q poa allocated-fees <operator-address>
```

**Example Response:**
```json
{
  "fees": {
    "fees": [
      {
        "denom": "token",
        "amount": "1000.500000000000000000"
      }
    ]
  }
}
```

---

### TotalPower

Get the total voting power across all validators.

**gRPC:** `cosmos.poa.v1.Query/TotalPower`
**REST:** `GET /cosmos/poa/v1/total_power`

**Request:**
```protobuf
message QueryTotalPowerRequest {}
```

**Response:**
```protobuf
message QueryTotalPowerResponse {
  int64 total_power = 1;
}
```

**CLI:**
```bash
simd q poa total-power
```

**Example Response:**
```json
{
  "total_power": "50000"
}
```

---

## Transaction Messages (Msg Service)

The Msg service handles state-changing operations.

### UpdateParams

Update module parameters (admin only).

**gRPC:** `cosmos.poa.v1.Msg/UpdateParams`

**Message:**
```protobuf
message MsgUpdateParams {
  Params params = 1;
  string admin = 2;  // Signer must be current admin
}
```

**Response:**
```protobuf
message MsgUpdateParamsResponse {}
```

**CLI:**
```bash
simd tx poa update-params \
  --admin <new-admin-address> \
  --from <current-admin> \
  --keyring-backend test \
  --chain-id <chain-id> \
  -y
```

**Authorization:** Only the current admin can execute this transaction.

---

### CreateValidator

Create a new validator with zero voting power (operator initiates, admin must activate).

**gRPC:** `cosmos.poa.v1.Msg/CreateValidator`

**Message:**
```protobuf
message MsgCreateValidator {
  google.protobuf.Any pub_key = 1;
  string moniker = 2;
  string description = 3;
  string operator_address = 4;  // Signer
}
```

**Response:**
```protobuf
message MsgCreateValidatorResponse {}
```

**CLI:**
```bash
simd tx poa create-validator \
  --pubkey <validator-pubkey> \
  --moniker "my-validator" \
  --description "Validator description" \
  --from <operator-account> \
  --keyring-backend test \
  --chain-id <chain-id> \
  -y
```

**Authorization:** Any account can create a validator, but it starts with power=0.

**Notes:**
- The validator will not participate in consensus until the admin updates its power to a non-zero value
- Public key must be a valid consensus public key (typically Ed25519)

---

### UpdateValidators

Update validator set (admin only). This is the primary mechanism for managing validators.

**gRPC:** `cosmos.poa.v1.Msg/UpdateValidators`

**Message:**
```protobuf
message MsgUpdateValidators {
  repeated Validator validators = 1;
  string admin = 2;  // Signer must be admin
}
```

**Response:**
```protobuf
message MsgUpdateValidatorsResponse {}
```

**CLI (inline):**
```bash
simd tx poa update-validators \
  --validator '{
    "pub_key": {
      "@type": "/cosmos.crypto.ed25519.PubKey",
      "key": "YUzyiqZzKN8BmLbl75gdXfbxQ2QtSYpPSwA85bZ3xuE="
    },
    "power": 10000
  }' \
  --validator '{
    "pub_key": {
      "@type": "/cosmos.crypto.ed25519.PubKey",
      "key": "lSR1GEByJtzgiuCevrWgcyBWjhQXjycsuzzIdf56Oa4="
    },
    "power": 0
  }' \
  --from account \
  --keyring-backend test \
  --chain-id <chain-id> \
  -y
```

**CLI (from file):**
```bash
simd tx poa update-validators validators.json \
  --from account \
  --keyring-backend test \
  --chain-id <chain-id> \
  -y
```

**File Format (validators.json):**
```json
[
  {
    "pub_key": {
      "@type": "/cosmos.crypto.ed25519.PubKey",
      "key": "YUzyiqZzKN8BmLbl75gdXfbxQ2QtSYpPSwA85bZ3xuE="
    },
    "power": 10000,
    "metadata": {
      "moniker": "validator-1",
      "description": "First validator",
      "operator_address": "cosmos1x0mm8rws8lm46xay3zyyznzr6lvu5um3kht0x7"
    }
  },
  {
    "pub_key": {
      "@type": "/cosmos.crypto.ed25519.PubKey",
      "key": "lSR1GEByJtzgiuCevrWgcyBWjhQXjycsuzzIdf56Oa4="
    },
    "power": 0
  }
]
```

**Authorization:** Only the admin can execute this transaction.

**Notes:**
- Can update multiple validators in a single transaction
- Setting `power: 0` removes a validator from the active set
- Changes propagate to CometBFT consensus in the next block
- Missing fields in metadata are preserved from existing state

---

### WithdrawFees

Withdraw accumulated fees to the operator's account.

**gRPC:** `cosmos.poa.v1.Msg/WithdrawFees`

**Message:**
```protobuf
message MsgWithdrawFees {
  string operator = 1;  // Signer
}
```

**Response:**
```protobuf
message MsgWithdrawFeesResponse {}
```

**CLI:**
```bash
simd tx poa withdraw-fees \
  --from <operator-account> \
  --keyring-backend test \
  --chain-id <chain-id> \
  -y
```

**Authorization:** Must be signed by the validator's operator address.

**Notes:**
- Transfers all accumulated fees to the operator's account
- Fees are denominated in the chain's native token(s)

---

## Common Use Cases

### 1. Query Current Admin

```bash
simd q poa params
```

### 2. List All Active Validators

```bash
simd q poa validators
```

### 3. Add a New Validator

**Step 1:** Operator creates the validator:
```bash
simd tx poa create-validator \
  --pubkey <pubkey> \
  --moniker "new-validator" \
  --from operator-account \
  --keyring-backend test \
  -y
```

**Step 2:** Admin activates with voting power:
```bash
simd tx poa update-validators \
  --validator '{
    "pub_key": {"@type": "/cosmos.crypto.ed25519.PubKey", "key": "..."},
    "power": 10000
  }' \
  --from admin \
  --keyring-backend test \
  -y
```

### 4. Change Validator Voting Power

```bash
simd tx poa update-validators \
  --validator '{
    "pub_key": {"@type": "/cosmos.crypto.ed25519.PubKey", "key": "..."},
    "power": 20000
  }' \
  --from admin \
  --keyring-backend test \
  -y
```

### 5. Remove a Validator

```bash
simd tx poa update-validators \
  --validator '{
    "pub_key": {"@type": "/cosmos.crypto.ed25519.PubKey", "key": "..."},
    "power": 0
  }' \
  --from admin \
  --keyring-backend test \
  -y
```

### 6. Withdraw Validator Fees

```bash
simd tx poa withdraw-fees \
  --from validator-operator \
  --keyring-backend test \
  -y
```

### 7. Transfer Admin Rights

```bash
simd tx poa update-params \
  --admin cosmos1newadminaddress... \
  --from current-admin \
  --keyring-backend test \
  -y
```

---

## REST API Endpoints

All query endpoints are available via REST:

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/cosmos/poa/v1/params` | Get module parameters |
| GET | `/cosmos/poa/v1/validator/{address}` | Get single validator |
| GET | `/cosmos/poa/v1/validators` | List all validators |
| GET | `/cosmos/poa/v1/allocated_fees/{operator_address}` | Get withdrawable fees |
| GET | `/cosmos/poa/v1/total_power` | Get total voting power |

**Example REST Query:**
```bash
curl http://localhost:1317/cosmos/poa/v1/validators
```

---

## Error Handling

Common error scenarios:

### Unauthorized Admin Action
**Error:** Transaction rejected
**Cause:** Non-admin attempted to call admin-only function
**Solution:** Ensure transaction is signed by the admin account

### Invalid Public Key
**Error:** Invalid validator public key
**Cause:** Malformed or wrong type of public key
**Solution:** Use Ed25519 public key in correct format

### Validator Not Found
**Error:** Validator does not exist
**Cause:** Querying non-existent validator
**Solution:** Verify validator address/public key

### Insufficient Fees
**Error:** No fees to withdraw
**Cause:** Validator has no accumulated fees
**Solution:** Wait for fees to accumulate from block rewards

---

## Integration Examples

### JavaScript/TypeScript (CosmJS)

```typescript
import { SigningStargateClient } from "@cosmjs/stargate";

// Query validators
const client = await StargateClient.connect("http://localhost:26657");
const response = await client.queryContractSmart(
  "cosmos.poa.v1.Query/Validators",
  {}
);

// Update validators (requires signing)
const signingClient = await SigningStargateClient.connectWithSigner(
  "http://localhost:26657",
  wallet
);

const msg = {
  typeUrl: "/cosmos.poa.v1.MsgUpdateValidators",
  value: {
    validators: [{
      pubKey: { typeUrl: "/cosmos.crypto.ed25519.PubKey", value: ... },
      power: 10000,
      metadata: { moniker: "validator-1", operatorAddress: "cosmos1..." }
    }],
    admin: "cosmos1adminaddress..."
  }
};

const result = await signingClient.signAndBroadcast(
  adminAddress,
  [msg],
  "auto"
);
```

### Python (cosmpy)

```python
from cosmpy.aerial.client import LedgerClient, NetworkConfig
from cosmpy.aerial.wallet import LocalWallet

# Create client
client = LedgerClient(NetworkConfig.fetchai_mainnet())

# Query validators
response = client.query_contract(
    "cosmos.poa.v1.Query/Validators",
    {}
)

print(response)
```

### Go

```go
import (
    "context"
    poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/types"
    "google.golang.org/grpc"
)

// Query client
conn, _ := grpc.Dial("localhost:9090", grpc.WithInsecure())
queryClient := poatypes.NewQueryClient(conn)

// Get validators
resp, err := queryClient.Validators(context.Background(), &poatypes.QueryValidatorsRequest{})
if err != nil {
    panic(err)
}

for _, val := range resp.Validators {
    fmt.Printf("Validator: %s, Power: %d\n", val.Metadata.Moniker, val.Power)
}
```

---

## Security Considerations

1. **Admin Key Security:** The admin private key has complete control over the validator set. Use hardware wallets or secure key management systems.

2. **Validator Public Keys:** Ensure validator public keys are correctly generated and stored securely.

3. **Power Distribution:** Consider the security implications of power concentration. Avoid giving a single validator >67% of total power.

4. **Operator Separation:** Use separate accounts for operator and admin roles to limit exposure.

5. **Fee Withdrawal:** Operators should regularly withdraw fees to prevent accumulation in the module.

---

## Appendix

### Public Key Formats

Ed25519 public keys should be base64-encoded:

```json
{
  "@type": "/cosmos.crypto.ed25519.PubKey",
  "key": "YUzyiqZzKN8BmLbl75gdXfbxQ2QtSYpPSwA85bZ3xuE="
}
```

### Address Formats

- **Operator Address:** Standard Cosmos SDK bech32 address (e.g., `cosmos1...`)
- **Consensus Address:** Can be derived from public key or use operator address for queries

### Power Units

- Voting power is represented as `int64`
- Total power affects block signing requirements (typically need >2/3 of total power for consensus)
- Zero power effectively removes a validator from the active set
s