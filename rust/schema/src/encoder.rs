use crate::value::{ArgValue, Value};

pub trait Encoder {
    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError>;
    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError>;
    // fn encode_struct();
    // fn encode_list_iterator<V, I: Iterator<Item=&V>>(&mut self, size: Option<usize>, );
    fn encode_list_slice<'a, V: ArgValue<'a>>(&mut self, x: &[V]) -> Result<(), EncodeError>;
}

pub enum EncodeError {}