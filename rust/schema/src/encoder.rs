//! Encoder trait and error type.
use crate::structs::{StructEncodeVisitor, StructType};
use crate::value::{SchemaValue};
use core::fmt::Display;
use ixc_message_api::AccountID;
use crate::list::ListEncodeVisitor;
use crate::r#enum::EnumType;

/// The trait that encoders must implement.
pub trait Encoder {
    /// Encode a `u32`.
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError>;
    /// Encode a `i32`.
    fn encode_i32(&mut self, x: i32) -> Result<(), EncodeError>;
    /// Encode a `u64`.
    fn encode_u64(&mut self, x: u64) -> Result<(), EncodeError>;
    /// Encode a `u128`.
    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError>;
    /// Encode a `str`.
    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError>;
    /// Encode a list.
    fn encode_list(&mut self, visitor: &dyn ListEncodeVisitor) -> Result<(), EncodeError>;
    /// Encode a struct.
    fn encode_struct(&mut self, visitor: &dyn StructEncodeVisitor, struct_type: &StructType) -> Result<(), EncodeError>;
    /// Encode an account ID.
    fn encode_account_id(&mut self, x: AccountID) -> Result<(), EncodeError>;
    /// Encode an enum value.
    fn encode_enum(&mut self, x: i32, enum_type: &EnumType) -> Result<(), EncodeError> {
        self.encode_i32(x)
    }
}

/// An encoding error.
#[derive(Debug)]
pub enum EncodeError {
    /// An unknown error occurred.
    UnknownError,
    /// The output buffer is out of space.
    OutOfSpace,
}