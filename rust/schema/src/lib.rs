//! This crate defines the basic traits and types for schema and encoding.
pub mod value;
pub mod types;
mod r#struct;

pub use r#struct::{StructCodec};
