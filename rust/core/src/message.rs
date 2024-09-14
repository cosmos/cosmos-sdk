use interchain_schema::StructCodec;

pub trait Message<const Mod: bool>: StructCodec {
    const SELECTOR: u128;
    type Response;
    type Error;
}