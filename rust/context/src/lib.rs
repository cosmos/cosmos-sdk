//! This crate provides an encoding agnostic implementation of a Context type which wraps
//! the basic RFC 003 message passing API along with an associated router and memory
//! management functions.

use interchain_message_api::{Address};

/// Context wraps a single message request (and possibly response as well) along with
/// the router callbacks necessary for making nested message calls.
pub struct Context {}

impl Context {
    /// This is the address of the account that is getting called.
    /// In a receiving account, this is the account's own address.
    pub fn address(&self) -> &Address {
        unimplemented!()
    }

    /// This is the address of the account which is making the message call.
    pub fn caller(&self) -> &Address {
        unimplemented!()
    }
}

/// Response is the type that should be used for message handler responses.
pub type Response<R, E=()> = Result<R, E>;