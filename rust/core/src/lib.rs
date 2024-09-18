//! This crate defines types and macros for constructing easy to use account and module implementations.
//! It integrates with the encoding layer but does not specify a state management framework.

#![cfg_attr(feature = "try_trait_v2", feature(try_trait_v2))]

#[cfg(feature = "std")]
extern crate alloc;

mod context;
mod response;
mod events;
mod message;
pub mod self_destruct;
pub mod handler;
pub mod resource;
mod on_create;
mod sync;

pub use context::Context;
pub use response::Response;
pub use events::EventBus;
pub use on_create::OnCreate;
