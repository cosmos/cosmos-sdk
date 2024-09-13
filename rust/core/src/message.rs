use interchain_schema::StructCodec;

pub trait Message<const Mod: bool>: StructCodec {
    type Response;
    type Error;
}