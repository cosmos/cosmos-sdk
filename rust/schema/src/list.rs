use bump_scope::BumpScope;
use crate::decoder::{DecodeError, Decoder};

pub trait ListVisitor<'a, T> {
    fn init(&mut self, len: usize, scope: &'a mut BumpScope<'a>) -> Result<(), DecodeError>;
    fn next<D: Decoder<'a>>(&mut self, decoder: &mut D) -> Result<(), DecodeError>;
}
