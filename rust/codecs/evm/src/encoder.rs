use alloy_sol_types::abi::encode;
use alloy_sol_types::SolValue;
use ixc_schema::encoder::EncodeError;
use ixc_schema::structs::StructEncodeVisitor;
use ixc_schema::value::Value;

struct Encoder {}

impl ixc_schema::encoder::Encoder for Encoder {
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError> {
        let bz = <u32 as SolValue>::abi_encode(&x);
        todo!()
    }

    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError> {
        let bz = <u128 as SolValue>::abi_encode(&x);
        todo!()
    }

    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_list_slice<'a, V: Value<'a>>(&mut self, xs: &[V]) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_struct<V: StructEncodeVisitor>(&mut self, visitor: &V) -> Result<(), EncodeError> {
        todo!()
    }
}