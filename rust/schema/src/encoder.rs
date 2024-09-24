use core::error::Error;
use core::fmt::{Display, Formatter};
use crate::r#struct::{StructDecodeVisitor, StructEncodeVisitor};
use crate::value::{Value, AbstractValue};

pub trait Encoder {
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError>;
    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError>;
    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError>;
    // fn encode_list_iterator<V, I: Iterator<Item=&V>>(&mut self, size: Option<usize>, );
    fn encode_list_slice<'a, V: Value<'a>>(&mut self, xs: &[V]) -> Result<(), EncodeError>;
    fn encode_struct<V: StructEncodeVisitor>(&mut self, visitor: &V) -> Result<(), EncodeError>;
}

#[derive(Debug)]
pub enum EncodeError {
    UnknownError,
    OutOfSpace
}