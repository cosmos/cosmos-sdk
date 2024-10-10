//! A codec for encoding and decoding Solidity ABI values.
mod encoder;
mod decoder;

use ixc_schema::buffer::WriterFactory;
use ixc_schema::codec::Codec;
use ixc_schema::encoder::EncodeError;
use ixc_schema::mem::MemoryManager;
use ixc_schema::value::SchemaValue;

/// A codec for encoding and decoding Solidity ABI values.
pub struct SolidityABICodec;

