#![doc = include_str!(concat!(env!("CARGO_MANIFEST_DIR"), "/README.md"))]

pub mod value;
pub mod types;
mod r#struct;
mod r#enum;
mod oneof;

pub use r#struct::{StructCodec};
pub use r#enum::{EnumCodec};
pub use oneof::{OneOfCodec};