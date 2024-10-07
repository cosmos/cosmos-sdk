use ixc_schema::encoder::EncodeError;
use ixc_schema::structs::{StructEncodeVisitor, StructType};
use integer_encoding::VarInt;
use ixc_schema::buffer::ReverseSliceWriter;

struct Encoder<'a> {
    writer: ReverseSliceWriter<'a>,
}

impl <'a> ixc_schema::encoder::Encoder for Encoder<'a> {
    fn encode_bool(&mut self, x: bool) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_u8(&mut self, x: u8) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_u16(&mut self, x: u16) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_u64(&mut self, x: u64) -> Result<(), EncodeError> {
        // fixed size buffer
        // <u64 as VarInt>::encode_var(x, &mut self.writer);
        todo!()
    }

    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_i8(&mut self, x: i8) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_i16(&mut self, x: i16) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_i32(&mut self, x: i32) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_i64(&mut self, x: i64) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_i128(&mut self, x: i128) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_bytes(&mut self, x: &[u8]) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_list(&mut self, visitor: &dyn ixc_schema::list::ListEncodeVisitor) -> Result<(), EncodeError> {
        todo!()
        // if it's a packed list type
        //  for each list item in reverse order
        //    encode element
        //  encode size
        //  encode tag
        // else
        //  for each list item in reverse order
        //    encode element
        //    encode tag
    }

    fn encode_struct(&mut self, visitor: &dyn StructEncodeVisitor, struct_type: &StructType) -> Result<(), EncodeError> {
        // for each field in reverse order
        //     encode field
        //     encode tag
        todo!()
    }

    fn encode_account_id(&mut self, x: ixc_message_api::AccountID) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_time(&mut self, x: simple_time::Time) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_duration(&mut self, x: simple_time::Duration) -> Result<(), EncodeError> {
        todo!()
    }
}