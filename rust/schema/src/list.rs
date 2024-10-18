//! Traits for encoding and decoding list types.
use allocator_api2::alloc::Allocator;
use allocator_api2::vec::Vec;
use crate::decoder::{DecodeError, Decoder};
use crate::encoder::{EncodeError, Encoder};
use crate::mem::MemoryManager;
use crate::value::SchemaValue;

/// A visitor for encoding list types.
pub trait ListEncodeVisitor {
    /// Get the size of the list if it is known or None if it is not known.
    fn size_hint(&self) -> Option<u32>;
    /// Encode the list.
    fn encode(&self, encoder: &mut dyn Encoder) -> Result<u32, EncodeError>;
    /// Encode the list in reverse order.
    fn encode_reverse(&self, encoder: &mut dyn Encoder) -> Result<u32, EncodeError>;
}

/// A visitor for decoding list types.
pub trait ListDecodeVisitor<'a> {
    /// Initialize the visitor with the length of the list.
    /// This method may or may not be called depending on whether the underlying
    /// encoding specifies the length of the list.
    fn init(&mut self, len: usize, scope: &'a MemoryManager) -> Result<(), DecodeError>;
    /// Decode the next element in the list.
    fn next(&mut self, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError>;
}

/// A builder for decoding Vec's with a specified allocator.
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

impl<'a, T: SchemaValue<'a>> ListDecodeVisitor<'a> for AllocatorVecBuilder<'a, T> {
    fn init(&mut self, len: usize, scope: &'a MemoryManager) -> Result<(), DecodeError> {
        self.get_xs(scope).reserve(len);
        Ok(())
    }

    fn next(&mut self, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        let mut state = T::DecodeState::default();
        T::visit_decode_state(&mut state, decoder)?;
        let value = T::finish_decode_state(state, decoder.mem_manager())?;
        self.get_xs(decoder.mem_manager()).push(value);
        Ok(())
    }
}

#[cfg(feature = "std")]
impl<'a, T: SchemaValue<'a>> ListDecodeVisitor<'a> for alloc::vec::Vec<T> {
    fn init(&mut self, len: usize, _scope: &'a MemoryManager) -> Result<(), DecodeError> {
        self.reserve(len);
        Ok(())
    }

    fn next(&mut self, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        let mut state = T::DecodeState::default();
        T::visit_decode_state(&mut state, decoder)?;
        self.push(T::finish_decode_state(state, decoder.mem_manager())?);
        Ok(())
    }
}

impl<'a, T: SchemaValue<'a>> ListEncodeVisitor for &'a [T] {
    fn size_hint(&self) -> Option<u32> {
        Some(self.len() as u32)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<u32, EncodeError> {
        for x in self.iter() {
            x.encode(encoder)?;
        }
        Ok(self.len() as u32)
    }

    fn encode_reverse(&self, encoder: &mut dyn Encoder) -> Result<u32, EncodeError> {
        for x in self.iter().rev() {
            x.encode(encoder)?;
        }
        Ok(self.len() as u32)
    }
}

#[cfg(feature = "std")]
impl<'a, T: SchemaValue<'a>> ListEncodeVisitor for alloc::vec::Vec<T> {
    fn size_hint(&self) -> Option<u32> {
        Some(self.len() as u32)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<u32, EncodeError> {
        for x in self.iter() {
            x.encode(encoder)?;
        }
        Ok(self.len() as u32)
    }

    fn encode_reverse(&self, encoder: &mut dyn Encoder) -> Result<u32, EncodeError> {
        for x in self.iter().rev() {
            x.encode(encoder)?;
        }
        Ok(self.len() as u32)
    }
}
