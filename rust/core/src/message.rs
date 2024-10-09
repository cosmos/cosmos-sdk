//! The Message trait for invoking messages dynamically.

use core::fmt::Debug;
use ixc_message_api::header::MessageSelector;
use ixc_schema::codec::Codec;
use ixc_schema::structs::StructSchema;
use ixc_schema::value::{OptionalValue, SchemaValue};

/// The Message trait for invoking messages dynamically.
pub trait Message<'a>: SchemaValue<'a> + StructSchema
// TODO required a sealed struct
{
    /// The message selector.
    const SELECTOR: MessageSelector;
    /// The optional response type.
    type Response<'b>: OptionalValue<'b>;
    /// The optional error type.
    type Error: Into<u8> + TryFrom<u8> + Debug;
    /// The codec to use for encoding and decoding the message.
    type Codec: Codec + Default;
}

/// Extract the response and error types from a Result.
/// Used internally in macros for building the Message implementation and ClientResult type.
pub trait ExtractResponseTypes {
    /// The response type.
    type Response;
    /// The error type.
    type Error;
    /// The client result type.
    type ClientResult;
}

impl<R, E: Debug + Into<u8> + TryFrom<u8>> ExtractResponseTypes for crate::Result<R, E> {
    type Response = R;
    type Error = E;
    type ClientResult = crate::result::ClientResult<R, E>;
}
