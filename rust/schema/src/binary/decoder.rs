use alloc::string::String;
use bump_scope::{BumpScope, BumpString};
use crate::decoder::DecodeError;
use crate::list::ListVisitor;
use crate::r#struct::StructDecodeVisitor;

struct Decoder<'a> {
    buf: &'a [u8],
    scope: &'a mut BumpScope<'a>,
}


impl <'a> Decoder<'a> {
    fn read_bytes(&mut self, size: usize) -> Result<&'a [u8], DecodeError> {
        if self.buf.len() < size {
            return Err(DecodeError::OutOfData);
        }
        let bz = &self.buf[0..size];
        self.buf = &self.buf[size..];
        Ok(bz)
    }
}

impl <'a> crate::decoder::Decoder<'a> for Decoder<'a> {
    fn decode_u32(&mut self) -> Result<i32, DecodeError> {
        let bz = self.read_bytes(4)?;
        Ok(i32::from_le_bytes(bz.try_into().unwrap()))
    }

    fn decode_u128(&mut self) -> Result<u128, DecodeError> {
        let bz = self.read_bytes(16)?;
        Ok(u128::from_le_bytes(bz.try_into().unwrap()))
    }

    fn decode_borrowed_str(&mut self) -> Result<Result<&'a str, BumpString<'a, 'a>>, DecodeError> {
        let bz = self.buf;
        self.buf = &[];
        Ok(Ok(core::str::from_utf8(bz).map_err(|_| DecodeError::InvalidData)?))
    }

    fn decode_owned_str(&mut self) -> Result<String, DecodeError> {
        let bz = self.buf;
        self.buf = &[];
        Ok(String::from_utf8(bz.to_vec()).map_err(|_| DecodeError::InvalidData)?)
    }

    fn decode_struct<V: StructDecodeVisitor<'a>>(&mut self, visitor: &mut V) -> Result<(), DecodeError> {
        // let mut i = 0;
        // let mut inner = InnerDecoder { outer: self };
        // for f in V::FIELDS {
        //     visitor.decode_field(i, &mut inner)?;
        //     i += 1;
        // }
        todo!();
        Ok(())
    }

    fn decode_list<T, V: ListVisitor<'a, T>>(&mut self, visitor: &mut V) -> Result<(), DecodeError> {
        // let size = self.decode_u32()? as usize;
        // visitor.init(size, &mut self.scope)?;
        // let mut inner = InnerDecoder { outer: self };
        // for _ in 0..size {
        //     visitor.next(&mut inner)?;
        // }
        todo!();
        Ok(())
    }

    fn scope(&self) -> &'a BumpScope<'a> {
        // &self.scope
        todo!();
    }
}

struct InnerDecoder<'a> {
    outer: &'a mut Decoder<'a>
}
impl <'a> crate::decoder::Decoder<'a> for InnerDecoder<'a> {
    fn decode_u32(&mut self) -> Result<i32, DecodeError> {
        self.outer.decode_u32()
    }

    fn decode_u128(&mut self) -> Result<u128, DecodeError> {
        self.outer.decode_u128()
    }

    fn decode_borrowed_str(&mut self) -> Result<Result<&'a str, BumpString<'a, 'a>>, DecodeError> {
        let size = self.decode_u32()? as usize;
        let bz = self.outer.read_bytes(size)?;
        Ok(Ok(core::str::from_utf8(bz).map_err(|_| DecodeError::InvalidData)?))
    }

    fn decode_owned_str(&mut self) -> Result<String, DecodeError> {
        let size = self.decode_u32()? as usize;
        let bz = self.outer.read_bytes(size)?;
        Ok(String::from_utf8(bz.to_vec()).map_err(|_| DecodeError::InvalidData)?)
    }

    fn decode_struct<V: StructDecodeVisitor<'a>>(&mut self, visitor: &mut V) -> Result<(), DecodeError> {
        // let size = self.decode_u32()? as usize;
        // let bz = self.outer.read_bytes(size)?;
        // let dec = Decoder { buf: bz, scope: self.outer.scope };
        todo!()
    }

    fn decode_list<T, V: ListVisitor<'a, T>>(&mut self, visitor: &mut V) -> Result<(), DecodeError> {
        todo!()
    }

    fn scope(&self) -> &'a BumpScope<'a> {
        todo!()
    }
}
