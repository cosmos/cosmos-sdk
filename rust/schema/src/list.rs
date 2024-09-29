use bump_scope::{Bump, BumpScope, BumpVec};
use crate::decoder::{DecodeError, Decoder};
use crate::mem::MemoryManager;
use crate::value::Value;

pub trait ListVisitor<'a, T> {
    fn init(&mut self, len: usize, scope: &MemoryManager<'a>) -> Result<(), DecodeError>;
    fn next<D: Decoder<'a>>(&mut self, decoder: &mut D) -> Result<(), DecodeError>;
}

pub struct SliceState<'a, T: Value<'a>> {
    pub(crate) xs: Option<BumpVec<'a, 'a, T>>,
}

impl <'a, T: Value<'a>> Default for SliceState<'a, T> {
    fn default() -> Self {
        Self {
            xs: None,
        }
    }
}

impl <'a, T: Value<'a>> SliceState<'a, T> {
    fn get_xs<'b>(&mut self, mem: &MemoryManager<'a>) -> &mut BumpVec<'a, 'a, T> {
        if self.xs.is_none() {
            self.xs = Some(mem.new_vec());
        }
        self.xs.as_mut().unwrap()
    }
}

impl <'a, T: Value<'a>> ListVisitor<'a, T> for SliceState<'a, T> {
    fn init(&mut self, len: usize, scope: &MemoryManager<'a>) -> Result<(), DecodeError> {
        self.get_xs(scope).reserve(len);
        Ok(())
    }

    fn next<D: Decoder<'a>>(&mut self, decoder: &mut D) -> Result<(), DecodeError> {
        let mut state = T::DecodeState::default();
        T::visit_decode_state(&mut state, decoder)?;
        let value = T::finish_decode_state(state, decoder.mem_manager())?;
        self.get_xs(decoder.mem_manager()).push(value);
        Ok(())
    }
}
