//! This crate defines types and macros for constructing easy to use account and module implementations.
//! It integrates with the encoding layer but does not specify a state management framework.
mod context;
mod response;
mod events;
mod message;
pub mod self_destruct;
pub mod handler;
pub mod resource;

pub use context::Context;
pub use response::Response;
pub use events::EventBus;
