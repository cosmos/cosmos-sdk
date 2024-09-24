use ixc_schema::StructCodec;
use ixc_schema::value::{ResponseValue, AbstractValue};

pub trait Message<const Mod: bool>: StructCodec + AbstractValue {
    const SELECTOR: u128;
    type Response: AbstractValue;
    type Error: AbstractValue;
}