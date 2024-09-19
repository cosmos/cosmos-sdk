use interchain_schema::StructCodec;
use interchain_schema::value::{ResponseValue, Value};

pub trait Message<const Mod: bool>: StructCodec + Value {
    const SELECTOR: u128;
    type Response: Value;
    type Error: Value;
}