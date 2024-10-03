//! **WARNING: This is an API preview! Most code won't work or even type check properly!**
//! This crate defines types and macros for constructing easy to use account and module implementations.
//! It integrates with the encoding layer but does not specify a state management framework.

#![doc = include_str!(concat!(env!("CARGO_MANIFEST_DIR"), "/README.md"))]
#![no_std]

#[cfg(feature = "std")]
extern crate alloc;

mod context;
mod events;
pub mod message;
pub mod account_api;
pub mod handler;
pub mod resource;
pub mod error;
pub mod routes;
pub mod low_level;

pub use context::Context;
pub use events::EventBus;
use ixc_schema::value::OptionalValue;
use crate::error::{Error, ErrorMessage};

/// The standard result type which wraps specific error types in the generic Error type
/// to include system error codes.
// pub type Result<R, E: OptionalValue<'static> = ErrorMessage> = core::result::Result<R, E>;
pub type Result<R> = core::result::Result<R, ()>;

