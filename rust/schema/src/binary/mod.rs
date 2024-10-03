//! Defines a codec for the native binary format.

use ixc_schema::binary::decoder::decode_value;
use crate::binary::encoder::encode_value;
use crate::buffer::WriterFactory;
use crate::codec::{Codec, ValueDecodeVisitor, ValueEncodeVisitor};
use crate::decoder::DecodeError;
use crate::encoder::EncodeError;
use crate::mem::MemoryManager;
use crate::state_object::{ObjectKey, ObjectValue};
use crate::value::SchemaValue;

pub(crate) mod encoder;
pub(crate) mod decoder;

/// A codec for encoding and decoding values using the native binary format.
#[derive(Default)]
pub struct NativeBinaryCodec;

impl Codec for NativeBinaryCodec {
    fn encode_value<'a>(&self, value: &dyn ValueEncodeVisitor, writer_factory: &'a dyn WriterFactory) -> Result<&'a [u8], EncodeError> {
        encode_value(value, writer_factory)
    }

    fn decode_value<'a>(&self, input: &'a [u8], memory_manager: &'a MemoryManager, visitor: &mut dyn ValueDecodeVisitor<'a>) -> Result<(), DecodeError> {
        decode_value(input, memory_manager, visitor)
    }
}
