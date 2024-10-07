mod encoder;
mod decoder;
mod wire;

use ixc_schema::buffer::WriterFactory;
use ixc_schema::codec::{Codec, ValueDecodeVisitor, ValueEncodeVisitor};
use ixc_schema::decoder::DecodeError;
use ixc_schema::encoder::EncodeError;
use ixc_schema::mem::MemoryManager;
use ixc_schema::value::SchemaValue;

pub struct ProtobufCodec;

impl Codec for ProtobufCodec {
    fn encode_value<'a>(&self, value: &dyn ValueEncodeVisitor, writer_factory: &'a dyn WriterFactory) -> Result<&'a [u8], EncodeError> {
        todo!()
    }

    fn decode_value<'a>(&self, input: &'a [u8], memory_manager: &'a MemoryManager, visitor: &mut dyn ValueDecodeVisitor<'a>) -> Result<(), DecodeError> {
        todo!()
    }
}