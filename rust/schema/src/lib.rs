#![doc = include_str!(concat!(env!("CARGO_MANIFEST_DIR"), "/README.md"))]
#![no_std]

#[cfg(feature = "std")]
extern crate alloc;

pub mod value;
pub mod types;
pub mod structs;
mod r#enum;
mod oneof;
pub mod state_object;
pub mod codec;
pub mod decoder;
mod list;
pub mod binary;
pub mod encoder;
mod kind;
mod field;
mod fields;
pub mod buffer;
pub mod mem;
mod mem2;
mod schema;
mod message;

pub use value::SchemaValue;