//! Defines a codec for the native binary format.
use crate::binary::encoder::encode_value;
use crate::buffer::WriterFactory;
use crate::codec::Codec;
use crate::decoder::DecodeError;
use crate::encoder::EncodeError;
use crate::mem::MemoryManager;
use crate::state_object::{ObjectKey, ObjectValue};
use crate::value::Value;

mod encoder;
mod decoder;
mod object_value;

/// A codec for encoding and decoding values using the native binary format.
pub struct NativeBinaryCodec;

impl Codec for NativeBinaryCodec {
    fn encode_value<'a, V: Value<'a>, F: WriterFactory>(value: &V, writer_factory: F) -> Result<F::Output, EncodeError> {
        encode_value(value, writer_factory)
    }

    fn decode_value<'a, V: Value<'a>>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<V, DecodeError> {
        decoder::decode_value(input, memory_manager)
    }
}

impl NativeBinaryCodec {
    pub fn encode_object_value<'a, V: ObjectValue, F: WriterFactory>(value: V::In<'a>, writer_factory: &F) -> Result<F::Output, EncodeError> {
        todo!()
    }

    pub fn decode_object_value<'a, V: ObjectValue>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<V::Out<'a>, DecodeError> {
        todo!()
    }
}