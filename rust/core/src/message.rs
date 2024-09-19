use ixc_schema::StructCodec;
use ixc_schema::value::{ResponseValue, Value};

pub trait Message<const Mod: bool>: StructCodec + Value {
    const SELECTOR: u128;
    type Response: Value;
    type Error: Value;
}