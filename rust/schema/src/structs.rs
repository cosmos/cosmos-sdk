//! Struct codec and schema traits.
use crate::decoder::{DecodeError, Decoder};
use crate::encoder::{EncodeError, Encoder};
use crate::field::Field;

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
/// use ixc_schema::StructCodec;
///
/// #[derive(StructCodec)]
/// pub struct MyStruct<'a> {
///   pub field1: u8,
///   pub field2: &'a str,
/// }
///
///
/// #[derive(StructCodec)]
/// pub struct MyStruct2 {
///   pub field1: simple_time::Time,
///   pub field2: ixc_message_api::Address,
/// }
/// ```
pub unsafe trait StructCodec {
    /// A dummy function for derived macro type checking.
    fn dummy(&self);
}

/// StructSchema is the trait that should be derived to define the schema of a struct.
pub unsafe trait StructSchema {
    /// The name of the struct.
    const NAME: &'static str;
    /// The fields of the struct.
    const FIELDS: &'static [Field<'static>];
    /// Whether the struct is sealed.
    const SEALED: bool;
}

/// StructDecodeVisitor is the trait that should be derived to decode a struct.
pub unsafe trait StructDecodeVisitor<'a>: StructSchema {
    /// Decode a field from the input data.
    fn decode_field<D: Decoder<'a>>(&mut self, index: usize, decoder: &mut D) -> Result<(), DecodeError>;
}

/// StructEncodeVisitor is the trait that should be derived to encode a struct.
pub unsafe trait StructEncodeVisitor: StructSchema {
    /// Encode a field to the output data.
    fn encode_field<E: Encoder>(&self, index: usize, encoder: &mut E) -> Result<(), EncodeError>;
}

/// StructType contains the schema of a struct.
pub struct StructType<'a> {
    /// The name of the struct.
    pub name: &'a str,
    /// The fields of the struct.
    pub fields: &'a [Field<'a>],
    /// Sealed indicates whether new fields can be added to the struct.
    /// If sealed is true, the struct is considered sealed and new fields cannot be added.
    pub sealed: bool,
}