use crate::encoder::{EncodeError};
use crate::value::ArgValue;

struct Encoder {

}

struct EncodeSizer {
    size: usize
}

impl crate::encoder::Encoder for EncodeSizer {
    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError> {
        self.size += 4;
        Ok(())
    }

    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError> {
        self.size += x.len();
        Ok(())
    }

    fn encode_list_slice<'a, V: ArgValue<'a>>(&mut self, x: &[V]) -> Result<(), EncodeError> {
        todo!()
    }
}

struct InnerEncoder {

}

struct InnerEncodeSizer {
    size: usize
}

impl crate::encoder::Encoder for InnerEncodeSizer {
    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError> {
        self.size += 4;
        Ok(())
    }

    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError> {
        self.size += 4;
        self.size += x.len();
        Ok(())
    }

    fn encode_list_slice<'a, V: ArgValue<'a>>(&mut self, xs: &[V]) -> Result<(), EncodeError> {
        self.size += 4;
        for x in xs.iter() {
           <V as ArgValue>::encode(x, self)?;
        }
        Ok(())
    }
}