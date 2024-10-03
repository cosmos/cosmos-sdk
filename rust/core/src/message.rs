use ixc_schema::codec::Codec;
use ixc_schema::structs::StructSchema;
use ixc_schema::value::{OptionalValue, SchemaValue};

pub trait Message<'a>: SchemaValue<'a> + StructSchema {
    const SELECTOR: u64;
    type Response<'b>: OptionalValue<'b>;
    type Error: OptionalValue<'static>;
    type Codec: Codec + Default;
}
