//! The codec trait.

use crate::decoder::Decoder;
use crate::buffer::{Writer, WriterFactory};
use crate::decoder::DecodeError;
use crate::encoder::{EncodeError, Encoder};
use crate::mem::MemoryManager;
use crate::value::SchemaValue;

/// Trait implemented by encoding protocols.
pub trait Codec {
    /// Encode a value.
    fn encode_value<'a>(&self, value: &dyn ValueEncodeVisitor, writer_factory: &'a dyn WriterFactory) -> Result<&'a [u8], EncodeError>;
    /// Decode a value.
    fn decode_value<'a>(&self, input: &'a [u8], memory_manager: &'a MemoryManager, visitor: &mut dyn ValueDecodeVisitor<'a>) -> Result<(), DecodeError>;
}

/// A visitor for encoding values. Unlike SchemaValue, this trait is object safe.
pub trait ValueEncodeVisitor {
    /// Visit the value.
    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError>;
}

impl <'a, T: SchemaValue<'a>> ValueEncodeVisitor for T {
    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        self.encode(encoder)
    }
}

/// A visitor for decoding values. Unlike SchemaValue, this trait is object safe.
pub trait ValueDecodeVisitor<'a> {
    /// Visit the value.
    fn decode(&mut self, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError>;
}

/// Decode a value.
pub fn decode_value<'a, V: SchemaValue<'a>>(codec: &dyn Codec, input: &'a [u8], mem: &'a MemoryManager) -> Result<V, DecodeError> {
    struct Visitor<'b, U:SchemaValue<'b>>(U::DecodeState);
    impl <'b, U:SchemaValue<'b>> ValueDecodeVisitor<'b> for Visitor<'b, U> {
        fn decode(&mut self, decoder: &mut dyn Decoder<'b>) -> Result<(), DecodeError> {
            U::visit_decode_state(&mut self.0, decoder)
        }
    }
    let mut visitor = Visitor::<V>(V::DecodeState::default());
    codec.decode_value(input, mem, &mut visitor)?;
    V::finish_decode_state(visitor.0, mem)
}