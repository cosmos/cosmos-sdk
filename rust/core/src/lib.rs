//! This crate defines types and macros for constructing easy to use account and module implementations.
//! It integrates with the encoding layer but does not specify a state management framework.

#[doc(inline)]
pub use interchain_message_api::{Address};
#[doc(inline)]
pub use interchain_context::{Context, Response};
#[doc(inline)]
pub use interchain_schema::{StructCodec};

#[cfg(feature = "core_macros")]
#[allow(unused_imports)]
#[macro_use]
extern crate interchain_core_macros;
#[cfg(feature = "core_macros")]
#[doc(inline)]
pub use interchain_core_macros::*;