//! This crate defines types and macros for constructing easy to use account and module implementations.
//! It integrates with the encoding layer but does not specify a state management framework.
mod context;
mod response;
mod events;
mod message;
pub mod self_destruct;
mod handler;
mod resource;

pub use context::Context;
pub use response::Response;
pub use events::EventBus;

#[doc(inline)]
pub use interchain_message_api::Address;
#[doc(inline)]
#[doc(inline)]
pub use interchain_schema::StructCodec;

#[cfg(feature = "core_macros")]
#[allow(unused_imports)]
#[macro_use]
extern crate interchain_core_macros;
#[cfg(feature = "core_macros")]
#[doc(inline)]
pub use interchain_core_macros::*;

#[cfg(feature = "schema_macros")]
#[allow(unused_imports)]
#[macro_use]
extern crate interchain_schema_macros;
#[cfg(feature = "schema_macros")]
#[doc(inline)]
pub use interchain_schema_macros::*;
