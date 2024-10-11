//! Struct codec and schema traits.
use crate::decoder::{DecodeError, Decoder};
use crate::encoder::{EncodeError, Encoder};
use crate::field::Field;
use crate::types::ReferenceableType;

/// StructSchema describes the schema of a struct.
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
