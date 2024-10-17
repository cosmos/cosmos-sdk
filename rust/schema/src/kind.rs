//! Field kinds.
use num_enum::{IntoPrimitive, TryFromPrimitive};

/// The basic type of a field.
#[non_exhaustive]
#[repr(u32)]
#[derive(TryFromPrimitive, IntoPrimitive, Debug, Clone, Copy, PartialEq, Eq)]
pub enum Kind {
    /// A string.
    String = 1,
    /// A byte array.
    Bytes = 2,
    /// A signed 8-bit integer.
    Int8 = 3,
    /// An unsigned 8-bit integer.
    Uint8 = 4,
    /// A signed 16-bit integer.
    Int16 = 5,
    /// An unsigned 16-bit integer.
    Uint16 = 6,
    /// A signed 32-bit integer.
    Int32 = 7,
    /// An unsigned 32-bit integer.
    Uint32 = 8,
    /// A signed 64-bit integer.
    Int64 = 9,
    /// An unsigned 64-bit integer.
    Uint64 = 10,
    /// A signed N-bye integer.
    IntN,
    /// An unsigned N-byte integer.
    UIntN,
    /// A decimal number.
    Decimal,
    /// A boolean.
    Bool,
    /// A timestamp with nano-second precision.
    Time,
    /// A duration with nano-second precision.
    Duration,
    /// A 32-bit floating point number.
    Float32,
    /// A 64-bit floating point number.
    Float64,
    /// An account ID.
    AccountID,
    /// An enumeration value.
    Enum,
    /// A JSON value.
    JSON,
    /// A struct value.
    Struct,
    /// A list value.
    List,
}