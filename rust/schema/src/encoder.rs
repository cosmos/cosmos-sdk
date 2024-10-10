//! Encoder trait and error type.

use core::error::Error;
use crate::structs::{StructEncodeVisitor, StructType};
use crate::value::{SchemaValue};
use core::fmt::Display;
use ixc_message_api::AccountID;
use ixc_message_api::code::{ErrorCode, SystemCode};
use ixc_schema::codec::ValueEncodeVisitor;
use crate::decoder::DecodeError;
use crate::list::ListEncodeVisitor;
use crate::r#enum::EnumType;

/// The trait that encoders must implement.
pub trait Encoder {
    /// Encode a `bool`.
    fn encode_bool(&mut self, x: bool) -> Result<(), EncodeError>;
    /// Encode a `u8`.
    fn encode_u8(&mut self, x: u8) -> Result<(), EncodeError>;
    /// Encode a `u16`.
    fn encode_u16(&mut self, x: u16) -> Result<(), EncodeError>;
    /// Encode a `u32`.
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError>;
    /// Encode a `u64`.
    fn encode_u64(&mut self, x: u64) -> Result<(), EncodeError>;
    /// Encode a `u128`.
    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError>;
    /// Encode an `i8`.
    fn encode_i8(&mut self, x: i8) -> Result<(), EncodeError>;
    /// Encode an `i16`.
    fn encode_i16(&mut self, x: i16) -> Result<(), EncodeError>;
    /// Encode an `i32`.
    fn encode_i32(&mut self, x: i32) -> Result<(), EncodeError>;
    /// Encode an `i64`.
    fn encode_i64(&mut self, x: i64) -> Result<(), EncodeError>;
    /// Encode an `i128`.
    fn encode_i128(&mut self, x: i128) -> Result<(), EncodeError>;
    /// Encode a `str`.
    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError>;
    /// Encode bytes.
    fn encode_bytes(&mut self, x: &[u8]) -> Result<(), EncodeError>;
    /// Encode a list.
    fn encode_list(&mut self, visitor: &dyn ListEncodeVisitor) -> Result<(), EncodeError>;
    /// Encode a struct.
    fn encode_struct(&mut self, visitor: &dyn StructEncodeVisitor, struct_type: &StructType) -> Result<(), EncodeError>;
    /// Encode a optional value.
    fn encode_option(&mut self, visitor: Option<&dyn ValueEncodeVisitor>) -> Result<(), EncodeError>;
    /// Encode an account ID.
    fn encode_account_id(&mut self, x: AccountID) -> Result<(), EncodeError>;
    /// Encode an enum value.
    fn encode_enum(&mut self, x: i32, enum_type: &EnumType) -> Result<(), EncodeError> {
        self.encode_i32(x)
    }
    /// Encode time.
    fn encode_time(&mut self, x: simple_time::Time) -> Result<(), EncodeError>;
    /// Encode duration.
    fn encode_duration(&mut self, x: simple_time::Duration) -> Result<(), EncodeError>;
}

/// An encoding error.
#[derive(Debug, Clone)]
#[non_exhaustive]
pub enum EncodeError {
    /// An unknown error occurred.
    UnknownError,
    /// The output buffer is out of space.
    OutOfSpace,
}

impl Display for EncodeError {
    fn fmt(&self, f: &mut core::fmt::Formatter<'_>) -> core::fmt::Result {
        match self {
            EncodeError::UnknownError => write!(f, "unknown error"),
            EncodeError::OutOfSpace => write!(f, "out of space"),
        }
    }
}

impl Error for EncodeError {}

impl From<EncodeError> for ErrorCode {
    fn from(value: EncodeError) -> Self {
        ErrorCode::SystemCode(SystemCode::EncodingError)
    }
}
