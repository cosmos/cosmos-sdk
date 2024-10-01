//! Encoder trait and error type.
use crate::structs::{StructEncodeVisitor, StructType};
use crate::value::{Value};
use core::fmt::Display;
use ixc_message_api::AccountID;

/// The trait that encoders must implement.
pub trait Encoder {
    /// Encode a `u32`.
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError>;
    /// Encode a `u64`.
    fn encode_u64(&mut self, x: u64) -> Result<(), EncodeError>;
    /// Encode a `u128`.
    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError>;
    /// Encode a `str`.
    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError>;
    // fn encode_list_iterator<V, I: Iterator<Item=&V>>(&mut self, size: Option<usize>, );
    /// Encode a list slice.
    fn encode_list_slice<'a, V: Value<'a>>(&mut self, xs: &[V]) -> Result<(), EncodeError>;
    /// Encode a struct.
    fn encode_struct<V: StructEncodeVisitor>(&mut self, visitor: &V, struct_type: &StructType) -> Result<(), EncodeError>;
    /// Encode an account ID.
    fn encode_account_id(&mut self, x: AccountID) -> Result<(), EncodeError>;
}

/// An encoding error.
#[derive(Debug)]
pub enum EncodeError {
    /// An unknown error occurred.
    UnknownError,
    /// The output buffer is out of space.
    OutOfSpace,
}