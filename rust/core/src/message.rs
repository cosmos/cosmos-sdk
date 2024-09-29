use ixc_schema::codec::Codec;
use ixc_schema::structs::StructSchema;
use ixc_schema::value::{ResponseValue, AbstractValue, Value};

pub trait Message<'a, const Mod: bool>: Value<'a> + StructSchema {
    const SELECTOR: u64;
    type Response: AbstractValue;
    type Error: AbstractValue;
    type Codec: Codec;
}
