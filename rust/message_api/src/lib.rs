//! **WARNING: This is an API preview! Most code won't work or even type check properly!**
//! This crate provides a low-level implementation of the Cosmos SDK RFC 003 message passing API.

mod header;
mod data_pointer;
pub mod packet;

/// A globally unique identifier for an account or message.
#[derive(Clone, Copy, Debug, PartialEq, Eq, PartialOrd, Ord, Hash, Default)]
pub struct Address {}

impl Address {
    /// Returns true if the address is empty and does not represent a valid account or message.
    pub const fn is_empty(&self) -> bool {
        unimplemented!()
    }
}