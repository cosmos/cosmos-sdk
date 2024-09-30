use ixc_schema::codec::Codec;
use ixc_schema::structs::StructSchema;
use ixc_schema::value::{ResponseValue, AbstractValue, Value};

pub trait Message<'a>: Value<'a> + StructSchema {
    const SELECTOR: u64;
    type Response<'b>: ResponseValue<'b>;
    type Error: ResponseValue<'static>;
    type Codec: Codec;
}
