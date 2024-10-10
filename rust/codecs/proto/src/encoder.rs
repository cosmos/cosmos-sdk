use ixc_schema::encoder::EncodeError;
use ixc_schema::structs::{StructEncodeVisitor, StructType};
use integer_encoding::VarInt;
use ixc_schema::buffer::{ReverseSliceWriter, Writer};
use ixc_schema::codec::ValueEncodeVisitor;
use crate::wire::{default_wire_info, WireInfo, WireType};

struct Encoder<'a> {
    writer: ReverseSliceWriter<'a>,
    cur_tag: u64,
    cur_wire_info: WireInfo,
    unpacked: bool,
    emit_defaults: bool,
}

impl<'a> ixc_schema::encoder::Encoder for Encoder<'a> {
    fn encode_bool(&mut self, x: bool) -> Result<(), EncodeError> {
        if !x && !self.emit_defaults {
            return Ok(());
        }
        self.writer.write(&[x as u8])
    }

    fn encode_u8(&mut self, x: u8) -> Result<(), EncodeError> {
        if x == 0 && !self.emit_defaults {
            return Ok(());
        }
        let mut buf = [0u8; 2];
        let n = <u8 as VarInt>::encode_var(x, &mut buf);
        self.writer.write(&buf[..n])
    }

    fn encode_u16(&mut self, x: u16) -> Result<(), EncodeError> {
        if x == 0 && !self.emit_defaults {
            return Ok(());
        }
        let mut buf = [0u8; 3];
        let n = <u16 as VarInt>::encode_var(x, &mut buf);
        self.writer.write(&buf[..n])
    }

    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError> {
        if x == 0 && !self.emit_defaults {
            return Ok(());
        }
        let mut buf = [0u8; 5];
        let n = <u32 as VarInt>::encode_var(x, &mut buf);
        self.writer.write(&buf[..n])
    }

    fn encode_u64(&mut self, x: u64) -> Result<(), EncodeError> {
        if x == 0 && !self.emit_defaults {
            return Ok(());
        }
        let mut buf = [0u8; 10];
        let n = <u64 as VarInt>::encode_var(x, &mut buf);
        self.writer.write(&buf[..n])
    }

    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError> {
        if x == 0 && !self.emit_defaults {
            return Ok(());
        }
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
        if !self.cur_wire_info.unpacked {
            // for each list item in reverse order
            //  encode element
        } else {
            // TODO: need to change the list encode visitor to encode one by one either forward or reverse
            // for each list item in reverse order
            //  encode element
            //  encode tag
        }
        todo!()
    }

    fn encode_struct(&mut self, visitor: &dyn StructEncodeVisitor, struct_type: &StructType) -> Result<(), EncodeError> {
        let mut num = 1;
        for field in struct_type.fields.iter().rev() {
            self.cur_wire_info = default_wire_info(field)?;
            self.cur_tag = crate::wire::encode_tag(num, self.cur_wire_info.wire_type)?;
            let end_pos = self.writer.pos();
            visitor.visit_field(field, self)?;
            // TODO deal with optional fields, approaches:
            // - pass an encoder instance or set some flag which doesn't emit defaults when we don't have an Option type
            //   and does emit them when we do
            // - change the visit_field signature to return a bool indicating whether the field had a default value
            //   and then we can decide whether to write the tag or not

            // if we have a regular field or a packed repeated field, we write the tag now
            // otherwise, the tag gets written many times during encode_list
            if !self.unpacked {
                if self.cur_wire_info.wire_type == WireType::LengthDelimited {
                    let start_pos = self.writer.pos();
                    let len = (end_pos - start_pos) as u64;
                    self.writer.write_u64(len)?;
                }
                self.writer.write_u64(self.cur_tag)?;
            }
            num += 1;
        }
        Ok(())
    }

    fn encode_option(&mut self, visitor: Option<&dyn ValueEncodeVisitor>) -> Result<(), EncodeError> {
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