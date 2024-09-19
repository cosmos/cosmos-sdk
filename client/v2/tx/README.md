The tx package provides a robust set of tools for building, signing, and managing transactions in a Cosmos SDK-based blockchain application.

## Overview

This package includes several key components:

1. Transaction Factory
2. Transaction Builder
3. Transaction Config
4. Transaction Encoder/Decoder
5. Signature Handling

## Architecture

```mermaid
graph TD
    A[Client] --> B[Factory]
    B --> C[TxBuilder]
    B --> D[TxConfig]
    D --> E[TxEncodingConfig]
    D --> F[TxSigningConfig]
    C --> G[Tx]
    G --> H[Encoder]
    G --> I[Decoder]
    F --> J[SignModeHandler]
    F --> K[SigningContext]
```

## Key Components

### TxConfig

`TxConfig` provides configuration for transaction handling, including:

- Encoding and decoding
- Sign mode handling
- Signature JSON marshaling/unmarshaling

```mermaid
classDiagram
    class TxConfig {
        <<interface>>
        TxEncodingConfig
        TxSigningConfig
        TxBuilderProvider
    }

    class TxEncodingConfig {
        <<interface>>
        TxEncoder() txEncoder
        TxDecoder() txDecoder
        TxJSONEncoder() txEncoder
        TxJSONDecoder() txDecoder
    }

    class TxSigningConfig {
        <<interface>>
        SignModeHandler() *signing.HandlerMap
        SigningContext() *signing.Context
        MarshalSignatureJSON([]Signature) ([]byte, error)
        UnmarshalSignatureJSON([]byte) ([]Signature, error)
    }

    class TxBuilderProvider {
        <<interface>>
    }

    class txConfig {
        TxBuilderProvider
        TxEncodingConfig
        TxSigningConfig
    }

    class defaultEncodingConfig {
        cdc codec.BinaryCodec
        decoder Decoder
        TxEncoder() txEncoder
        TxDecoder() txDecoder
        TxJSONEncoder() txEncoder
        TxJSONDecoder() txDecoder
    }

    class defaultTxSigningConfig {
        signingCtx *signing.Context
        handlerMap *signing.HandlerMap
        cdc codec.BinaryCodec
        SignModeHandler() *signing.HandlerMap
        SigningContext() *signing.Context
        MarshalSignatureJSON([]Signature) ([]byte, error)
        UnmarshalSignatureJSON([]byte) ([]Signature, error)
    }

    TxConfig <|-- txConfig
    TxEncodingConfig <|.. defaultEncodingConfig
    TxSigningConfig <|.. defaultTxSigningConfig
    txConfig *-- TxBuilderProvider
    txConfig *-- defaultEncodingConfig
    txConfig *-- defaultTxSigningConfig
```

### TxBuilder

`TxBuilder` is responsible for constructing the transaction.

```mermaid
classDiagram
    class TxBuilder {
        <<interface>>
        GetTx() (Tx, error)
        GetSigningTxData() (*signing.TxData, error)
        SetMsgs(...transaction.Msg) error
        SetMemo(string)
        SetFeeAmount([]*base.Coin)
        SetGasLimit(uint64)
        SetTimeoutHeight(uint64)
        SetFeePayer(string) error
        SetFeeGranter(string) error
        SetUnordered(bool)
        SetSignatures(...Signature) error
    }

    class ExtendedTxBuilder {
        <<interface>>
        +SetExtensionOptions(...*gogoany.Any)
    }

    class txBuilder {
        addressCodec address.Codec
        decoder Decoder
        codec codec.BinaryCodec
        msgs []transaction.Msg
        timeoutHeight uint64
        granter []byte
        payer []byte
        unordered bool
        memo string
        gasLimit uint64
        fees []*base.Coin
        signerInfos []*apitx.SignerInfo
        signatures [][]byte
        extensionOptions []*anypb.Any
        nonCriticalExtensionOptions []*anypb.Any
        GetTx() (Tx, error)
        GetSigningTxData() (*signing.TxData, error)
        SetMsgs(...transaction.Msg) error
        SetMemo(string)
        SetFeeAmount([]*base.Coin)
        SetGasLimit(uint64)
        SetTimeoutHeight(uint64)
        SetFeePayer(string) error
        SetFeeGranter(string) error
        SetUnordered(bool)
        SetSignatures(...Signature) error
        getTx() (*wrappedTx, error)
        getFee() (*apitx.Fee, error)
    }

    class TxBuilderProvider {
        <<interface>>
        NewTxBuilder() TxBuilder
    }

    class BuilderProvider {
        addressCodec address.Codec
        decoder Decoder
        codec codec.BinaryCodec
        NewTxBuilder() TxBuilder
    }

    TxBuilder <|.. txBuilder : implements
    ExtendedTxBuilder <|.. txBuilder : implements
    TxBuilderProvider <|.. BuilderProvider : implements
    BuilderProvider ..> txBuilder : creates
```

### Factory

The `Factory` is the main entry point for creating and managing transactions. It handles:

- Account preparation
- Gas calculation
- Transaction simulation
- Unsigned transaction building

```mermaid
classDiagram
    class Factory {
        keybase keyring.Keyring
        cdc codec.BinaryCodec
        accountRetriever account.AccountRetriever
        ac address.Codec
        conn gogogrpc.ClientConn
        txConfig TxConfig
        txParams TxParameters

        NewFactory(keybase, cdc, accRetriever, txConfig, ac, conn, parameters) Factory
        Prepare() error
        BuildUnsignedTx(msgs ...transaction.Msg) (TxBuilder, error)
        BuildsSignedTx(ctx context.Context, msgs ...transaction.Msg) (Tx, error)
        calculateGas(msgs ...transaction.Msg) error
        Simulate(msgs ...transaction.Msg) (*apitx.SimulateResponse, uint64, error)
        UnsignedTxString(msgs ...transaction.Msg) (string, error)
        BuildSimTx(msgs ...transaction.Msg) ([]byte, error)
        sign(ctx context.Context, txBuilder TxBuilder, overwriteSig bool) (Tx, error)
        WithGas(gas uint64)
        WithSequence(sequence uint64)
        WithAccountNumber(accnum uint64)
        preprocessTx(keyname string, builder TxBuilder) error
        accountNumber() uint64
        sequence() uint64
        GgasAdjustment() float64
        keyring() keyring.Keyring
        simulateAndExecute() bool
        signMode() apitxsigning.SignMode
        getSimPK() (cryptotypes.PubKey, error)
        getSimSignatureData(pk cryptotypes.PubKey) SignatureData
        getSignBytesAdapter(ctx context.Context, signerData signing.SignerData, builder TxBuilder) ([]byte, error)
    }

    class TxParameters {
        <<struct>>
    }

    class TxConfig {
        <<interface>>
    }

    class TxBuilder {
        <<interface>>
    }

    Factory *-- TxParameters
    Factory *-- TxConfig
    Factory ..> TxBuilder : creates and uses
```

### Encoder/Decoder

The package includes functions for encoding and decoding transactions in both binary and JSON formats.

```mermaid
classDiagram
    class Decoder {
        <<interface>>
        Decode(txBytes []byte) (*txdecode.DecodedTx, error)
    }

    class txDecoder {
        <<function>>
        decode(txBytes []byte) (Tx, error)
    }

    class txEncoder {
        <<function>>
        encode(tx Tx) ([]byte, error)
    }

    class EncoderUtils {
        <<utility>>
        decodeTx(cdc codec.BinaryCodec, decoder Decoder) txDecoder
        encodeTx(tx Tx) ([]byte, error)
        decodeJsonTx(cdc codec.BinaryCodec, decoder Decoder) txDecoder
        encodeJsonTx(tx Tx) ([]byte, error)
        protoTxBytes(tx *txv1beta1.Tx) ([]byte, error)
    }

    class MarshalOptions {
        <<utility>>
        Deterministic bool
    }

    class JSONMarshalOptions {
        <<utility>>
        Indent string
        UseProtoNames bool
        UseEnumNumbers bool
    }

    Decoder <.. EncoderUtils : uses
    txDecoder <.. EncoderUtils : creates
    txEncoder <.. EncoderUtils : implements
    EncoderUtils ..> MarshalOptions : uses
    EncoderUtils ..> JSONMarshalOptions : uses
```

### Sequence Diagrams

#### Generate Aux Signer Data
```mermaid
sequenceDiagram
    participant User
    participant GenerateOrBroadcastTxCLI
    participant generateAuxSignerData
    participant makeAuxSignerData
    participant AuxTxBuilder
    participant ctx.PrintProto

    User->>GenerateOrBroadcastTxCLI: Call with isAux flag
    GenerateOrBroadcastTxCLI->>generateAuxSignerData: Call

    generateAuxSignerData->>makeAuxSignerData: Call
    makeAuxSignerData->>AuxTxBuilder: NewAuxTxBuilder()
    
    makeAuxSignerData->>AuxTxBuilder: SetAddress(f.txParams.fromAddress)
    
    alt f.txParams.offline
        makeAuxSignerData->>AuxTxBuilder: SetAccountNumber(f.AccountNumber())
        makeAuxSignerData->>AuxTxBuilder: SetSequence(f.Sequence())
    else
        makeAuxSignerData->>f.accountRetriever: GetAccountNumberSequence()
        makeAuxSignerData->>AuxTxBuilder: SetAccountNumber(accNum)
        makeAuxSignerData->>AuxTxBuilder: SetSequence(seq)
    end
    
    makeAuxSignerData->>AuxTxBuilder: SetMsgs(msgs...)
    makeAuxSignerData->>AuxTxBuilder: SetSignMode(f.SignMode())
    
    makeAuxSignerData->>f.keybase: GetPubKey(f.txParams.fromName)
    makeAuxSignerData->>AuxTxBuilder: SetPubKey(pubKey)
    
    makeAuxSignerData->>AuxTxBuilder: SetChainID(f.txParams.chainID)
    makeAuxSignerData->>AuxTxBuilder: GetSignBytes()
    
    makeAuxSignerData->>f.keybase: Sign(f.txParams.fromName, signBz, f.SignMode())
    makeAuxSignerData->>AuxTxBuilder: SetSignature(sig)
    
    makeAuxSignerData->>AuxTxBuilder: GetAuxSignerData()
    AuxTxBuilder-->>makeAuxSignerData: Return AuxSignerData
    makeAuxSignerData-->>generateAuxSignerData: Return AuxSignerData
    
    generateAuxSignerData->>ctx.PrintProto: Print AuxSignerData
    ctx.PrintProto-->>GenerateOrBroadcastTxCLI: Return result
    GenerateOrBroadcastTxCLI-->>User: Return result
```
#### Generate Only
```mermaid
sequenceDiagram
    participant User
    participant GenerateOrBroadcastTxCLI
    participant generateOnly
    participant Factory
    participant TxBuilder
    participant ctx.PrintString

    User->>GenerateOrBroadcastTxCLI: Call with generateOnly flag
    GenerateOrBroadcastTxCLI->>generateOnly: Call

    generateOnly->>Factory: Prepare()
    alt Error in Prepare
        Factory-->>generateOnly: Return error
        generateOnly-->>GenerateOrBroadcastTxCLI: Return error
        GenerateOrBroadcastTxCLI-->>User: Return error
    end

    generateOnly->>Factory: UnsignedTxString(msgs...)
    Factory->>Factory: SimulateAndExecute()
    alt SimulateAndExecute is true
        Factory->>Factory: calculateGas(msgs...)
        Factory->>Factory: Simulate(msgs...)
        Factory->>Factory: WithGas(adjusted)
    end

    Factory->>Factory: BuildUnsignedTx(msgs...)
    Factory->>TxBuilder: NewTxBuilder()
    Factory->>TxBuilder: SetMsgs(msgs...)
    Factory->>TxBuilder: SetMemo(f.txParams.memo)
    Factory->>TxBuilder: SetFeeAmount(fees)
    Factory->>TxBuilder: SetGasLimit(f.txParams.gas)
    Factory->>TxBuilder: SetFeeGranter(f.txParams.feeGranter)
    Factory->>TxBuilder: SetFeePayer(f.txParams.feePayer)
    Factory->>TxBuilder: SetTimeoutHeight(f.txParams.timeoutHeight)

    Factory->>TxBuilder: GetTx()
    Factory->>Factory: txConfig.TxJSONEncoder()
    Factory->>Factory: encoder(tx)

    Factory-->>generateOnly: Return unsigned tx string
    generateOnly->>ctx.PrintString: Print unsigned tx string
    ctx.PrintString-->>generateOnly: Return result
    generateOnly-->>GenerateOrBroadcastTxCLI: Return result
    GenerateOrBroadcastTxCLI-->>User: Return result
```

#### DryRun
```mermaid
sequenceDiagram
    participant User
    participant GenerateOrBroadcastTxCLI
    participant dryRun
    participant Factory
    participant os.Stderr

    User->>GenerateOrBroadcastTxCLI: Call with dryRun flag
    GenerateOrBroadcastTxCLI->>dryRun: Call

    dryRun->>Factory: Check txParams.offline
    alt txParams.offline is true
        Factory-->>dryRun: Return error (cannot use offline mode)
        dryRun-->>GenerateOrBroadcastTxCLI: Return error
        GenerateOrBroadcastTxCLI-->>User: Return error
    end

    dryRun->>Factory: Prepare()
    alt Error in Prepare
        Factory-->>dryRun: Return error
        dryRun-->>GenerateOrBroadcastTxCLI: Return error
        GenerateOrBroadcastTxCLI-->>User: Return error
    end

    dryRun->>Factory: Simulate(msgs...)
    Factory->>Factory: BuildSimTx(msgs...)
    Factory->>Factory: BuildUnsignedTx(msgs...)
    Factory->>Factory: getSimPK()
    Factory->>Factory: getSimSignatureData(pk)
    Factory->>Factory: SetSignatures(sig)
    Factory->>Factory: TxEncoder()(tx)
    
    Factory->>Factory: txConfig.SignModeHandler().GetSignBytes()
    Factory->>Factory: keybase.Sign()
    Factory->>Factory: SetSignatures()

    Factory->>ServiceClient: Simulate(context.Background(), &apitx.SimulateRequest{})
    ServiceClient->>Factory: Return result
    
    Factory-->>dryRun: Return (simulation, gas, error)
    alt Error in Simulate
        dryRun-->>GenerateOrBroadcastTxCLI: Return error
        GenerateOrBroadcastTxCLI-->>User: Return error
    end

    dryRun->>os.Stderr: Fprintf(GasEstimateResponse{GasEstimate: gas})
    os.Stderr-->>dryRun: Return result
    dryRun-->>GenerateOrBroadcastTxCLI: Return result
    GenerateOrBroadcastTxCLI-->>User: Return result
```

#### Generate and Broadcast Tx
```mermaid
sequenceDiagram
    participant User
    participant GenerateOrBroadcastTxCLI
    participant BroadcastTx
    participant Factory
    participant TxBuilder
    participant clientCtx

    User->>GenerateOrBroadcastTxCLI: Call
    GenerateOrBroadcastTxCLI->>BroadcastTx: Call

    BroadcastTx->>Factory: Prepare()
    alt Error in Prepare
        Factory-->>BroadcastTx: Return error
        BroadcastTx-->>GenerateOrBroadcastTxCLI: Return error
        GenerateOrBroadcastTxCLI-->>User: Return error
    end

    BroadcastTx->>Factory: SimulateAndExecute()
    alt SimulateAndExecute is true
        BroadcastTx->>Factory: calculateGas(msgs...)
        Factory->>Factory: Simulate(msgs...)
        Factory->>Factory: WithGas(adjusted)
    end

    BroadcastTx->>Factory: BuildUnsignedTx(msgs...)
    Factory->>TxBuilder: NewTxBuilder()
    Factory->>TxBuilder: SetMsgs(msgs...)
    Factory->>TxBuilder: SetMemo(memo)
    Factory->>TxBuilder: SetFeeAmount(fees)
    Factory->>TxBuilder: SetGasLimit(gas)
    Factory->>TxBuilder: SetFeeGranter(feeGranter)
    Factory->>TxBuilder: SetFeePayer(feePayer)
    Factory->>TxBuilder: SetTimeoutHeight(timeoutHeight)

    alt !clientCtx.SkipConfirm
        BroadcastTx->>TxBuilder: GetTx()
        BroadcastTx->>Factory: txConfig.TxJSONEncoder()
        BroadcastTx->>clientCtx: PrintRaw(txBytes)
        BroadcastTx->>clientCtx: Input.GetConfirmation()
        alt Not confirmed
            BroadcastTx-->>GenerateOrBroadcastTxCLI: Return error
            GenerateOrBroadcastTxCLI-->>User: Return error
        end
    end

    BroadcastTx->>Factory: sign(ctx, builder, true)
    Factory->>Factory: keybase.GetPubKey(fromName)
    Factory->>Factory: getSignBytesAdapter()
    Factory->>Factory: keybase.Sign(fromName, bytesToSign, signMode)
    Factory->>TxBuilder: SetSignatures(sig)
    Factory->>TxBuilder: GetTx()

    BroadcastTx->>Factory: txConfig.TxEncoder()
    BroadcastTx->>clientCtx: BroadcastTx(txBytes)

    alt Error in BroadcastTx
        clientCtx-->>BroadcastTx: Return error
        BroadcastTx-->>GenerateOrBroadcastTxCLI: Return error
        GenerateOrBroadcastTxCLI-->>User: Return error
    end

    BroadcastTx->>clientCtx: OutputTx(res)
    clientCtx-->>BroadcastTx: Return result
    BroadcastTx-->>GenerateOrBroadcastTxCLI: Return result
    GenerateOrBroadcastTxCLI-->>User: Return result
```

## Usage

To use the `tx` package, typically you would:

1. Create a `Factory`
2. Prepare the account
3. Build an unsigned transaction
4. Simulate the transaction (optional)
5. Sign the transaction
6. Broadcast the transaction

Here's a simplified example:

```go
factory, _ := NewFactory(...)
factory.Prepare()
txBuilder, _ := factory.BuildUnsignedTx(msgs...)
factory.Sign(ctx, txBuilder, true)
txBytes, _ := factory.txConfig.TxEncoder()(txBuilder.GetTx())
// Broadcast txBytes
```