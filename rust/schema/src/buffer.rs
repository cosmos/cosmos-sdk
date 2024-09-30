//! Buffer utilities for encoding and decoding.
use bump_scope::{BumpScope, BumpBox, BumpVec};
use crate::encoder::EncodeError;
use crate::mem::MemoryManager;

/// A factory for creating writers.
pub trait WriterFactory {
    /// The type of output produced by the writer.
    type Output;
    /// Create a new reverse writer.
    fn new_reverse(&self, size: usize) -> impl Writer<Output=Self::Output>;
}

/// A writer that writes bytes slices in the order specified when it was created.
pub trait Writer {
    /// The type of output produced by the writer.
    type Output;
    /// Write bytes to the buffer.
    fn write(&mut self, bytes: &[u8]) -> Result<(), EncodeError>;
    /// Get the current position in the buffer.
    fn pos(&self) -> usize;
    /// Finish writing and return the output.
    fn finish(self) -> Result<Self::Output, EncodeError>;
}

impl<'a> WriterFactory for &'a MemoryManager {
    type Output = &'a [u8];

    fn new_reverse(&self, size: usize) -> impl Writer<Output=Self::Output> {
        let b = self.bump.alloc_slice_fill(size, 0);
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

impl<'a> Writer for ReverseSliceWriter<'a> {
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