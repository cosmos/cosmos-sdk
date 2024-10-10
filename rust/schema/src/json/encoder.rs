use ixc_message_api::AccountID;
use simple_time::{Duration, Time};
use crate::codec::ValueEncodeVisitor;
use crate::encoder::EncodeError;
use crate::list::ListEncodeVisitor;
use crate::structs::{StructEncodeVisitor, StructType};

struct Encoder {

}

impl crate::encoder::Encoder for Encoder {
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_i32(&mut self, x: i32) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_u64(&mut self, x: u64) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_list(&mut self, visitor: &dyn ListEncodeVisitor) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_struct(&mut self, visitor: &dyn StructEncodeVisitor, struct_type: &StructType) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_account_id(&mut self, x: AccountID) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_bool(&mut self, x: bool) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_u8(&mut self, x: u8) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_u16(&mut self, x: u16) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_i8(&mut self, x: i8) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_i16(&mut self, x: i16) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_i64(&mut self, x: i64) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_i128(&mut self, x: i128) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_bytes(&mut self, x: &[u8]) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_time(&mut self, x: Time) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_duration(&mut self, x: Duration) -> Result<(), EncodeError> {
        todo!()
    }

    fn encode_option(&mut self, visitor: Option<&dyn ValueEncodeVisitor>) -> Result<(), EncodeError> {
        todo!()
    }
}