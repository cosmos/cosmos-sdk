use bump_scope::{Bump, BumpScope, BumpVec};
use crate::decoder::{DecodeError, Decoder};
use crate::value::Value;

pub trait ListVisitor<'a, T> {
    fn init(&mut self, len: usize, scope: &mut BumpScope<'a>) -> Result<(), DecodeError>;
    fn next<D: Decoder<'a>>(&mut self, decoder: &mut D) -> Result<(), DecodeError>;
}

pub struct SliceState<'a, T: Value<'a>> {
    xs: Option<BumpVec<'a, 'a, T>>,
    mem_handles: Option<BumpVec<'a, 'a, T::MemoryHandle>>,
}

impl <'a, T: Value<'a>> Default for SliceState<'a, T> {
    fn default() -> Self {
        Self {
            xs: None,
            mem_handles: None,
        }
    }
}

impl <'a, T: Value<'a>> SliceState<'a, T> {
    fn get_xs<'b>(&mut self, scope: &'b BumpScope<'a>) -> &mut BumpVec<'b, 'a, T> {
        // self.xs.get_or_insert_with(|| BumpVec::new_in(scope))
        todo!()
    }

    fn get_mem_handles<'b>(&mut self, scope: &'b BumpScope<'a>) -> &mut BumpVec<'b, 'a, T::MemoryHandle> {
        // self.mem_handles.get_or_insert_with(|| BumpVec::new_in(scope))
        todo!()
    }
}

impl <'a, T: Value<'a>> ListVisitor<'a, T> for SliceState<'a, T> {
    fn init(&mut self, len: usize, scope: &mut BumpScope<'a>) -> Result<(), DecodeError> {
        self.get_xs(scope).reserve(len);
        Ok(())
    }

    fn next<D: Decoder<'a>>(&mut self, decoder: &mut D) -> Result<(), DecodeError> {
        let mut state = T::DecodeState::default();
        T::visit_decode_state(&mut state, decoder)?;
        let (value, mem_handle) = T::finish_decode_state(state)?;
        self.get_xs(decoder.scope()).push(value);
        if let Some(mem_handle) = mem_handle {
            self.get_mem_handles(decoder.scope()).push(mem_handle);
        }
        Ok(())
    }
}
