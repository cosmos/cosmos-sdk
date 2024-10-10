use ixc_schema::codec::ValueDecodeVisitor;
use ixc_schema::decoder::DecodeError;
use ixc_schema::mem::MemoryManager;
use ixc_schema::structs::{StructDecodeVisitor, StructType};

struct Decoder<'a> {
    data: &'a [u8],
}

impl <'a> ixc_schema::decoder::Decoder<'a> for Decoder<'a> {
    fn decode_bool(&mut self) -> Result<bool, DecodeError> {
        todo!()
    }

    fn decode_u8(&mut self) -> Result<u8, DecodeError> {
        todo!()
    }

    fn decode_u16(&mut self) -> Result<u16, DecodeError> {
        todo!()
    }

    fn decode_u32(&mut self) -> Result<u32, DecodeError> {
        todo!()
    }

    fn decode_u64(&mut self) -> Result<u64, DecodeError> {
        todo!()
    }

    fn decode_u128(&mut self) -> Result<u128, DecodeError> {
        todo!()
    }

    fn decode_i8(&mut self) -> Result<i8, DecodeError> {
        todo!()
    }

    fn decode_i16(&mut self) -> Result<i16, DecodeError> {
        todo!()
    }

    fn decode_i32(&mut self) -> Result<i32, DecodeError> {
        todo!()
    }

    fn decode_i64(&mut self) -> Result<i64, DecodeError> {
        todo!()
    }

    fn decode_i128(&mut self) -> Result<i128, DecodeError> {
        todo!()
    }

    fn decode_borrowed_str(&mut self) -> Result<&'a str, DecodeError> {
        todo!()
    }

    fn decode_owned_str(&mut self) -> Result<String, DecodeError> {
        todo!()
    }

    fn decode_borrowed_bytes(&mut self) -> Result<&'a [u8], DecodeError> {
        todo!()
    }

    fn decode_owned_bytes(&mut self) -> Result<Vec<u8>, DecodeError> {
        todo!()
    }

    fn decode_struct(&mut self, visitor: &mut dyn StructDecodeVisitor<'a>, struct_type: &StructType) -> Result<(), DecodeError> {
        todo!()
    }

    fn decode_list(&mut self, visitor: &mut dyn ixc_schema::list::ListDecodeVisitor<'a>) -> Result<(), DecodeError> {
        // if it's a packed tag
        //  decode size
        //  for each list item
        //    decode element
        // else
        //  decode next element
        todo!()
    }

    fn decode_option(&mut self, visitor: &mut dyn ValueDecodeVisitor<'a>) -> Result<bool, DecodeError> {
        todo!()
    }

    fn decode_account_id(&mut self) -> Result<ixc_message_api::AccountID, DecodeError> {
        todo!()
    }

    fn decode_time(&mut self) -> Result<simple_time::Time, DecodeError> {
        todo!()
    }

    fn decode_duration(&mut self) -> Result<simple_time::Duration, DecodeError> {
        todo!()
    }

    fn mem_manager(&self) -> &'a MemoryManager {
        todo!()
    }
}