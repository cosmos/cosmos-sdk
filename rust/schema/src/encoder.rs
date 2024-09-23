use crate::r#struct::{StructDecodeVisitor, StructEncodeVisitor};
use crate::value::{ArgValue, Value};

pub trait Encoder {
    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError>;
    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError>;
    // fn encode_list_iterator<V, I: Iterator<Item=&V>>(&mut self, size: Option<usize>, );
    fn encode_list_slice<'a, V: ArgValue<'a>>(&mut self, xs: &[V]) -> Result<(), EncodeError>;
    fn encode_struct<'a, V: StructEncodeVisitor>(&mut self, visitor: &V) -> Result<(), EncodeError>;
}

pub enum EncodeError {
    UnknownError
}