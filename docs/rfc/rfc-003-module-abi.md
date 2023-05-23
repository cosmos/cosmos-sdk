# RFC 003: Language-independent Module Semantics & ABI

## Changelog

* 2023-03-10: Initial draft

## Background

This document aims to propose a scope of work that introduces support for different programming languages within the Cosmos SDK. Unlike traditional RFCs, it will also provide a high-level overview of why this proposal is being made and how it will impact users.

Currently, the Cosmos SDK allows developers to write modules in Golang, which has been a popular choice. However, the broader Cosmos ecosystem has witnessed the widespread adoption of other languages through virtual machines (VMs) such as Cosmwasm (Wasm), Agoric (JS), Polaris (Solidity), and Ethermint (Solidity). Additionally, within the wider blockchain ecosystem, Rust and various Rust subset VMs have gained significant traction.

Given this landscape, the question arises: should the Cosmos SDK support different languages and offer better support for their usage? To address this question, it's helpful to view the SDK as consisting of two distinct layers: the kernel space and the user space. In this perspective, modules represent the kernel space, while VMs constitute the user space. When discussing the support for different languages, both spaces need to be considered.

The primary objective is to enable modules to be written in different programming languages, allowing them to serve as core components of a node. Simultaneously, users should have the ability to write VMs in Rust and connect them to a general interface that provides similar, albeit limited, functionality as a module. This approach ensures flexibility in language choice for both core development and user customization within the Cosmos ecosystem.

## Proposal

This RFC specifies a way for first-class Cosmos SDK modules to be written in other
languages such as Rust. It aims to provide as much parity as possible between modules defined in Go and those defined
in other languages.

This ABI assumes a C-calling convention and would allow a module to be loaded via `cgo`. An additional binding layer
could be added on top to use a module via WASM, but that will be discussed in a separate RFC that follows up on this
one.

Module developers would not interact with this ABI directly but would instead have more user friendly, type safe tooling
built on top of it. In Rust, this might look like a combination of generated code (for protobuf types) and macros for
describing providers. The Rust ABI will be specified in an RFC that follows up on this one.

This specification is related to [Cosmos Proto Zero-Copy Encoding DRAFT Spec](/6ICE-uQpTDSF1PxiJbPUXw) which specifies
how messages specified in .proto files can be passed between modules written in different languages and VMs with zero or
almost zero overhead. This encoding is used extensively by 

## Memory Management

All functions are defined such that the caller allocates and frees any shared memory. Callees are not expected to return any memory buffers that the caller would then need to free. Callback functions are used in cases where the callee is expected to pass memory to the caller. In any callback function, it is expected that memory is read and copied if the data needs to be retained.

## Entrypoints

### Protobuf `FileDescriptor`s

Before doing any other sort of initialization the host runtime and guest modules must synchronize their `FileDescriptor` sets to make sure that all descriptors are compatible in terms of encoding.

This would be exposed through the function `cosmos_read_file_descriptors`:

**Rust**

```rust!
#[no_mangle]
pub extern fn cosmos_read_file_descriptors(callback: extern fn(size: u32, gzipped_bytes: *const u8) -> i32) -> i32
```

When the host calls `cosmos_read_file_descriptors` the module should call `callback` with the size and gzipped `FileDescriptor` bytes for each `FileDescriptor`. A non-zero return code for the function and callback is used to indicate an error.

### Module Registration

A single code unit may register one or more modules. Each module is identified by a unique protobuf configuration type as with Cosmos SDK `appconfig` modules, ex. `cosmos.bank.module.v1.Module`. Module registration happens similarly to `depinject` registration in the SDK where each module can define one or more provider functions. Each provider is described statically by a `ProviderInfo` (see below) so that the framework can build the dependency graph in the correct order.

The `cosmos_register_modules` function is called to register module providers. Each module should call the `register` function that is passed in with:

* a `ModuleInfo` message (as described below) encoded with [Cosmos Proto Zero-Copy Encoding DRAFT Spec](/6ICE-uQpTDSF1PxiJbPUXw) is passed as `module_info_data` with its size set to `module_info_size`
* `providers`: an array of provider callback functions whose size must be equal to the number of providers described in `ModuleInfo`

A provider function is called with:

* `config_data`:
* `inputs`: an array of inputs which are cast to the type expected by each input type with size equal to that specified in `ProviderInfo`
* `register_output`: a callback function that should be called once for each output specified in `ProviderInfo` with a value of corresponding to expected output type
* `err`: a 64kb buffer where an error can be optionally written to. In the case of an error, the return code should be non-zero

```rust!
#[no_mangle]
pub extern fn cosmos_register_modules(register: RegisterFn
) -> i32

type RegisterFn = extern fn(module_info_size: u32, module_info_data: *const u8, providers: *const ProviderFn) -> i32

type ProviderFn = extern fn(config_size: u32, config_data: *const u8, inputs: *const Void, register_output: RegisterOutputFn, err: *u8) -> i32

type RegisterOutputFn = extern fn(output: *const Void)
```

```protobuf!
message ModuleInfo {
  // module_config_type is the fully qualified name of the module config type.
  string module_config_type = 1;
  // providers describes the inputs and outputs of each provider function.
  repeated ProviderInfo providers = 2;
}

message ProviderInfo {
  repeated inputs = 1;
  repeated outputs = 2;
}

message Input {
  Type type = 1;
  bool optional = 2;
  
  message Type {
    oneof type {
      // service is the fully-qualified name of the service.
      string service = 1;
    }
  }
}

message Output {
  Type type = 1;
  
  message Type {
    oneof type {
      // service is the fully-qualified name of the service.
      string service = 1;
    }
  }
}
```

A set of functions and macros would be provided in languages like Rust to do this registration in a type safe way.

## Services

This design takes a "services all the way down" architecture approach. All functions that a module invokes and provides
are described by `service` definitions in .proto files. These correspond to the `service` provider input and output
type.

The following types of services are supported:

* transaction services which are annotated by the `cosmos.msg.v1.service` annotation and contain state machine logic that can be invoked via transactions and inter-module calls
* query services which are un-annotated and are executed in a read-only context. Only service methods annotated with `cosmos.query.v1.module_query_safe` can be called from other modules
* internal services which can only be called from other modules annotated with `cosmos.msg.v1.internal_service` (TBD). Internal services also receive the name of the calling module in their context pointer to do authentication. In this way, even a service like storage could be managed in this way because it knows which module called it.
* app module services to support things like genesis, begin and end blockers. These services can be defined once per module and are annotated with `cosmos.app.services.v1.module_service` option.

Refer to the [`cosmos.app.services.v1`](../../proto/cosmos/app/services/v1) protobuf package to see the core set of
services that are supported by the framework.

### Service Methods

All transaction and query service methods must be unary methods (meaning that they don't support streaming). App module services might use client or server side streaming (but not bidirectional streaming) to support things like genesis import and export and store iterators (we need to verify that store iterators can adequately be represented with streaming). Internal services may also support client or server side streaming.

### Service ABI

A service when passed as an input or output to a `ProviderFn` is represented as an array of method function pointers where the size and order of the array corresponds to the size and order of methods declared in the code unit's protobuf `FileDescriptor` for that service.

### Method ABI

### Unary Methods

The same ABI can be used for modules to both implement service methods and call service methods as a client.

**Rust**

```rust!
type UnaryMethodFn = extern fn(ctx: *const u8, req_size: u32, req: *const u8, res_size: u32, res: *mut u8) -> i32;
```

Unary methods takes the following parameters:

* `ctx`: an opaque context pointer
* `req`: a pointer to memory representing the request data encoded with [Cosmos Proto Zero-Copy Encoding DRAFT Spec](/6ICE-uQpTDSF1PxiJbPUXw) whose size is `req_size`
* `res` a pointer to a block of memory with size `res_size` where the response data is to be written, either:
    * the response encoded with [Cosmos Proto Zero-Copy Encoding DRAFT Spec](/6ICE-uQpTDSF1PxiJbPUXw)
    * a null-terminated string describing an error

The return value for unary methods is a 32-bit signed integer where a non-zero value represents the number of bytes written to the response buffer and a negative value represents a pre-defined error code.

The response memory buffer is allocated and freed by the caller so that service implementations do not have to deal with returning a memory buffer that the caller needs to free.

By default, all response buffers should be 64kb in length as specified in [Cosmos Proto Zero-Copy Encoding DRAFT Spec](/6ICE-uQpTDSF1PxiJbPUXw) unless annotations on protobuf types (such as a `max_size` validation criteria) allow the caller to safely calculate a larger or smaller value. Until such as specification exists, 64kb buffers should be used and should be sufficient for all applications that do not involve storing byte code. Modules that need larger buffer sizes for such applications should be written as native go modules for now.

### Client and Server Streaming Methods

TODO

## Abandoned Ideas (Optional)


## References

## Discussion
