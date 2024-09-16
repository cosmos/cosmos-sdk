//! This crate defines the basic traits and types for schema and encoding.
pub mod value;
pub mod types;
mod r#struct;

pub use r#struct::{StructCodec};

#[cfg(feature = "macros")]
#[allow(unused_imports)]
#[macro_use]
extern crate interchain_schema_macros;
#[cfg(feature = "macros")]
#[doc(inline)]
pub use interchain_schema_macros::*;
