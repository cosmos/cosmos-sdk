//! The Message trait for invoking messages dynamically.

use allocator_api2::alloc::Allocator;
use ixc_message_api::handler::{HandlerError, HostBackend};
use ixc_message_api::header::MessageSelector;
use ixc_message_api::packet::MessagePacket;
use ixc_schema::codec::{decode_value, Codec};
use ixc_schema::mem::MemoryManager;
use ixc_schema::structs::StructSchema;
use ixc_schema::value::{OptionalValue, SchemaValue};
use crate::Context;

/// The Message trait for invoking messages dynamically.
pub trait Message<'a>: SchemaValue<'a> + StructSchema {
    /// The message selector.
    const SELECTOR: MessageSelector;
    /// The optional response type.
    type Response<'b>: OptionalValue<'b>;
    /// The optional error type.
    type Error: OptionalValue<'static>;
    /// The codec to use for encoding and decoding the message.
    type Codec: Codec + Default;
}

/// Extract the response and error types from a Result.
/// Used internally for building the Message trait with a macro.
pub trait ExtractResponseTypes {
    /// The response type.
    type Response;
    /// The error type.
    type Error;
}

impl<R, E> ExtractResponseTypes for core::result::Result<R, E> {
    type Response = R;
    type Error = E;
}
