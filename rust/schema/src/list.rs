use allocator_api2::alloc::Allocator;
use allocator_api2::vec::Vec;
use crate::decoder::{DecodeError, Decoder};
use crate::mem::MemoryManager;
use crate::value::SchemaValue;

pub trait ListVisitor<'a, T> {
    fn init(&mut self, len: usize, scope: &'a MemoryManager) -> Result<(), DecodeError>;
    fn next<D: Decoder<'a>>(&mut self, decoder: &mut D) -> Result<(), DecodeError>;
}

pub struct AllocatorVecBuilder<'a, T: SchemaValue<'a>> {
    pub(crate) xs: Option<Vec<T, &'a dyn Allocator>>,
}

impl<'a, T: SchemaValue<'a>> Default for AllocatorVecBuilder<'a, T> {
    fn default() -> Self { Self { xs: None } }
}

impl<'a, T: SchemaValue<'a>> AllocatorVecBuilder<'a, T> {
    fn get_xs<'b>(&mut self, mem: &'a MemoryManager) -> &mut Vec<T, &'a dyn Allocator> {
        if self.xs.is_none() {
            self.xs = Some(Vec::new_in(mem));
        }
        self.xs.as_mut().unwrap()
    }
}

impl<'a, T: SchemaValue<'a>> ListVisitor<'a, T> for AllocatorVecBuilder<'a, T> {
    fn init(&mut self, len: usize, scope: &'a MemoryManager) -> Result<(), DecodeError> {
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

#[cfg(feature = "std")]
impl<'a, T: SchemaValue<'a>> ListVisitor<'a, T> for alloc::vec::Vec<T> {
    fn init(&mut self, len: usize, _scope: &'a MemoryManager) -> Result<(), DecodeError> {
        self.reserve(len);
        Ok(())
    }

    fn next<D: Decoder<'a>>(&mut self, decoder: &mut D) -> Result<(), DecodeError> {
        let mut state = T::DecodeState::default();
        T::visit_decode_state(&mut state, decoder)?;
        self.push(T::finish_decode_state(state, decoder.mem_manager())?);
        Ok(())
    }
}
