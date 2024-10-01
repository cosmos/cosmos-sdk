//! Struct codec and schema traits.
use crate::decoder::{DecodeError, Decoder};
use crate::encoder::{EncodeError, Encoder};
use crate::field::Field;
use crate::types::ReferenceableType;

/// StructCodec is the trait that should be derived to encode and decode a struct.
///
/// It should generally be used in conjunction with the `#[derive(StructCodec)]` attribute
/// attached to a struct definition.
/// It is unsafe to implement this trait manually, because the compiler cannot
///  guarantee correct implementations.
///
/// Any struct which contains fields which implement [`value::MaybeBorrowed`] can be derived
/// to implement this trait.
/// Structs and their fields may optionally contain a single lifetime parameter, in which
/// case decoded values will be borrowed from the input data wherever possible.
///
/// Example:
/// ```
/// use ixc_schema::SchemaValue;
///
/// #[derive(SchemaValue)]
/// pub struct MyStruct<'a> {
///   pub field1: u8,
///   pub field2: &'a str,
/// }
///
///
/// #[derive(SchemaValue)]
/// pub struct MyStruct2 {
///   pub field1: simple_time::Time,
///   pub field2: ixc_message_api::AccountID,
/// }
/// ```
/// StructSchema is the trait that should be derived to define the schema of a struct.
pub unsafe trait StructSchema: ReferenceableType {
    /// The schema of the struct.
    const STRUCT_TYPE: StructType<'static>;
}

/// StructDecodeVisitor is the trait that should be derived to decode a struct.
pub unsafe trait StructDecodeVisitor<'a> {
    /// Decode a field from the input data.
    fn decode_field(&mut self, index: usize, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError>;
}

/// StructEncodeVisitor is the trait that should be derived to encode a struct.
pub unsafe trait StructEncodeVisitor {
    /// Encode a field to the output data.
    fn encode_field(&self, index: usize, encoder: &mut dyn Encoder) -> Result<(), EncodeError>;
}

/// StructType contains the schema of a struct.
#[derive(Debug, Clone, Eq, PartialEq)]
pub struct StructType<'a> {
    /// The name of the struct.
    pub name: &'a str,
    /// The fields of the struct.
    pub fields: &'a [Field<'a>],
    /// Sealed indicates whether new fields can be added to the struct.
    /// If sealed is true, the struct is considered sealed and new fields cannot be added.
    pub sealed: bool,
}
