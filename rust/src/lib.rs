#![doc = include_str!(concat!(env!("CARGO_MANIFEST_DIR"), "/README.md"))]


#[doc(inline)]
pub use interchain_core::Context;
#[doc(inline)]
pub use interchain_core::Response;
#[doc(inline)]
pub use interchain_core::EventBus;

#[doc(inline)]
pub use interchain_message_api::Address;
#[doc(inline)]
pub use interchain_schema::StructCodec;
#[doc(inline)]
pub use state_objects::*;
#[doc(inline)]
pub use simple_time::{Time, Duration};

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
