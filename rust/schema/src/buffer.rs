//! Buffer utilities for encoding and decoding.

use crate::encoder::EncodeError;
use allocator_api2::alloc::Allocator;
use core::alloc::Layout;
use core::ptr::slice_from_raw_parts_mut;

/// A factory for creating writers.
pub trait WriterFactory {
    /// The type of output produced by the writer.
    type Output;
    /// Create a new reverse writer.
    fn new_reverse(&self, size: usize) -> Result<impl Writer<Output=Self::Output>, EncodeError>;
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

impl<'a> WriterFactory for &'a dyn Allocator {
    type Output = &'a [u8];

    fn new_reverse(&self, size: usize) -> Result<impl Writer<Output=Self::Output>, EncodeError> {
        unsafe {
            let ptr = self.allocate_zeroed(
                Layout::from_size_align_unchecked(size, 1)
            ).map_err(|_| EncodeError::OutOfSpace)?;
            Ok(ReverseSliceWriter {
                buf: &mut *ptr.as_ptr(),
                pos: size,
            })
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