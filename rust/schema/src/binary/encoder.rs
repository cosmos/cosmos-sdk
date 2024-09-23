use crate::encoder::{EncodeError};
use crate::r#struct::{StructDecodeVisitor, StructEncodeVisitor};
use crate::value::ArgValue;

struct Encoder<W> {
    writer: W
}

trait Writer {
    fn write(&mut self, bytes: &[u8]) -> Result<(), EncodeError>;
}

impl <W: Writer> crate::encoder::Encoder for Encoder<W> {
    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError> {
        self.writer.write(&x.to_le_bytes())
    }

    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError> {
        self.writer.write(x.as_bytes())
    }

    fn encode_list_slice<'a, V: ArgValue<'a>>(&mut self, x: &[V]) -> Result<(), EncodeError> {
        self.writer.write(&(x.len() as u32).to_le_bytes())?;
        for v in x.iter() {
            <V as ArgValue>::encode(v, self)?;
        }
        Ok(())
    }

    fn encode_struct<'a, V: StructEncodeVisitor>(&mut self, visitor: &V) -> Result<(), EncodeError> {
        // let mut i = 0;
        // let mut inner = InnerEncoder::<W> { outer: self };
        // for f in V::FIELDS {
        //     visitor.encode_field(i, &mut inner)?;
        //     i += 1;
        // }
        todo!();
        Ok(())
    }
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

    fn encode_list_slice<'a, V: ArgValue<'a>>(&mut self, xs: &[V]) -> Result<(), EncodeError> {
        self.size += 4;
        for x in xs.iter() {
            <V as ArgValue>::encode(x, self)?;
        }
        Ok(())
    }

    fn encode_struct<'a, V: StructEncodeVisitor>(&mut self, visitor: &V) -> Result<(), EncodeError> {
        todo!()
    }
}

struct InnerEncoder<W> {
    outer: Encoder<W>
}

impl <W: Writer> crate::encoder::Encoder for InnerEncoder<W> {
    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError> {
        self.outer.encode_u128(x)
    }

    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError> {
        self.outer.writer.write(&(x.len() as u32).to_le_bytes())?;
        self.outer.encode_str(x)
    }

    fn encode_list_slice<'a, V: ArgValue<'a>>(&mut self, x: &[V]) -> Result<(), EncodeError> {
        // TODO: prefix with length of actual encoded data
        self.outer.writer.write(&(x.len() as u32).to_le_bytes())?;
        self.outer.encode_list_slice(x)
    }

    fn encode_struct<'a, V: StructEncodeVisitor>(&mut self, visitor: &V) -> Result<(), EncodeError> {
        todo!()
    }
}

struct InnerEncodeSizer {
    outer: EncodeSizer
}

impl crate::encoder::Encoder for InnerEncodeSizer {
    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError> {
        self.outer.encode_u128(x)
    }

    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError> {
        self.outer.size += 4;
        self.outer.encode_str(x)
    }

    fn encode_list_slice<'a, V: ArgValue<'a>>(&mut self, xs: &[V]) -> Result<(), EncodeError> {
        self.outer.size += 4;
        self.outer.encode_list_slice(xs)
    }

    fn encode_struct<'a, V: StructEncodeVisitor>(&mut self, visitor: &V) -> Result<(), EncodeError> {
        todo!()
    }
}
