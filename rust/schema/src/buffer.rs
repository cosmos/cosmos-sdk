//! Buffer utilities for encoding and decoding.
use bump_scope::{BumpScope, BumpBox};
use crate::encoder::EncodeError;

/// A factory for creating writers.
pub trait WriterFactory {
    /// The type of output produced by the writer.
    type Output;
    /// Create a new reverse writer.
    fn new_reverse(&self, size: usize) -> impl ReverseWriter<Output=Self::Output>;
}

/// A writer that writes bytes slices starting from the end of a buffer.
pub trait ReverseWriter {
    /// The type of output produced by the writer.
    type Output;
    /// Write bytes to the end of the buffer.
    fn write(&mut self, bytes: &[u8]) -> Result<(), EncodeError>;
    /// Get the current position in the buffer.
    fn pos(&self) -> usize;
    /// Finish writing and return the output.
    fn finish(self) -> Result<Self::Output, EncodeError>;
}

impl<'a> WriterFactory for BumpScope<'a> {
    type Output = &'a [u8];
    fn new_reverse(&self, size: usize) -> impl ReverseWriter<Output=Self::Output> {
        let b = self.alloc_slice_fill(size, 0);
        ReverseSliceWriter {
            buf: b.into_mut(),
            pos: size,
        }
    }
}


struct ReverseSliceWriter<'a> {
    buf: &'a mut [u8],
    pos: usize,
}

impl<'a> ReverseWriter for ReverseSliceWriter<'a> {
    type Output = &'a [u8];

    fn write(&mut self, bytes: &[u8]) -> Result<(), EncodeError> {
        if self.pos < bytes.len() {
            return Err(EncodeError::OutOfSpace);
        }
        self.pos -= bytes.len();
        self.buf[self.pos..self.pos + bytes.len()].copy_from_slice(bytes);
        Ok(())
    }

    fn pos(&self) -> usize {
        self.pos
    }

    fn finish(self) -> Result<&'a [u8], EncodeError> {
        Ok(&self.buf[self.pos..])
    }
}