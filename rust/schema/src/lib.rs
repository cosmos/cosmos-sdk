#![doc = include_str!(concat!(env!("CARGO_MANIFEST_DIR"), "/README.md"))]
#![no_std]

#[cfg(feature = "std")]
extern crate alloc;

pub mod value;
pub mod types;
mod r#struct;
mod r#enum;
mod oneof;
pub mod state_object;
mod codec;
mod decoder;
mod list;
mod binary;
mod encoder;
mod proto;
mod kind;
mod field;
mod fields;
mod buffer;
mod mem;

pub use r#struct::{StructCodec};
pub use r#enum::{EnumCodec};
pub use oneof::{OneOfCodec};