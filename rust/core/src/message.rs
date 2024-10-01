use ixc_schema::codec::Codec;
use ixc_schema::structs::StructSchema;
use ixc_schema::value::{ResponseValue, SchemaValue};

pub trait Message<'a>: SchemaValue<'a> + StructSchema {
    const SELECTOR: u64;
    type Response<'b>: ResponseValue<'b>;
    type Error: ResponseValue<'static>;
    type Codec: Codec;
}
