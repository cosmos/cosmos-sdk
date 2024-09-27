mod encoder;
mod decoder;

use ixc_schema::buffer::WriterFactory;
use ixc_schema::codec::Codec;
use ixc_schema::encoder::EncodeError;
use ixc_schema::mem::MemoryManager;
use ixc_schema::value::Value;

pub struct SolidityABICodec;

impl Codec for SolidityABICodec {
    fn encode_value<'a, V: Value<'a>, F: WriterFactory>(value: &V, writer_factory: &F) -> Result<F::Output, EncodeError> {
        todo!()
    }

    fn decode_value<'b, 'a: 'b, V: Value<'a>>(input: &'a [u8], memory_manager: &'b MemoryManager<'a, 'a>) -> Result<V, DecodeError> {
        todo!()
    }
}
