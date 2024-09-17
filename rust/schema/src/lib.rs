//! This crate defines the basic traits and types for schema and encoding.
pub mod value;
pub mod types;
mod r#struct;
mod r#enum;
mod oneof;

pub use r#struct::{StructCodec};
pub use r#enum::{EnumCodec};
pub use oneof::{OneOfCodec};