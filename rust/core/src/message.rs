use ixc_schema::codec::Codec;
use ixc_schema::StructCodec;
use ixc_schema::value::{ResponseValue, AbstractValue, Value};

pub trait Message<'a, const Mod: bool>: Value<'a> + AbstractValue {
    const SELECTOR: [u8; 16];
    type Response: AbstractValue;
    type Error: AbstractValue;
    type Codec: Codec;
}