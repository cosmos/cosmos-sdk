use crate::encoder::{EncodeError, Encoder};
use crate::value::ArgValue;

pub struct BinaryEncoder {

}

struct EncodeSizer {
    size: usize
}

impl Encoder for EncodeSizer {
    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError> {
        self.size += 4;
        Ok(())
    }

    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError> {
        self.size += 4;
        Ok(())
    }

    fn encode_list_slice<'a, V: ArgValue<'a>>(&mut self, x: &[V]) -> Result<(), EncodeError> {
        todo!()
    }
}