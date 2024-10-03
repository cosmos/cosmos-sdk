//! Buffer utilities for encoding and decoding.

use crate::encoder::EncodeError;
use allocator_api2::alloc::Allocator;
use core::alloc::Layout;
use core::ptr::slice_from_raw_parts_mut;
use crate::decoder::DecodeError;

/// A factory for creating writers.
pub trait WriterFactory {
    /// Create a new reverse writer.
    fn new_reverse(&self, size: usize) -> Result<ReverseSliceWriter, EncodeError>;
}

/// A writer that writes bytes slices in the order specified when it was created.
pub trait Writer {
    /// Write bytes to the buffer.
    fn write(&mut self, bytes: &[u8]) -> Result<(), EncodeError>;
    /// Get the current position in the buffer.
    fn pos(&self) -> usize;
}


impl<'a> WriterFactory for &'a dyn Allocator {
    fn new_reverse(&self, size: usize) -> Result<ReverseSliceWriter, EncodeError> {
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

/// A writer that writes bytes slices in reverse order.
pub struct ReverseSliceWriter<'a> {
    buf: &'a mut [u8],
    pos: usize,
}

impl<'a> Writer for ReverseSliceWriter<'a> {
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
}

impl<'a> ReverseSliceWriter<'a> {
    /// Finish writing and return the buffer.
    pub fn finish(self) -> &'a [u8] {
        &self.buf[self.pos..]
    }
}

/// A buffer reader.
pub trait Reader<'a> {
    /// Read a slice of bytes from the buffer and update the remaining length.
    fn read_bytes(&mut self, size: usize) -> Result<&'a [u8], DecodeError>;

    /// Check if the buffer has been completely read and return an error if not.
    fn is_done(&self) -> Result<(), DecodeError>;
}

impl <'a> Reader<'a> for &'a [u8] {
    fn read_bytes(&mut self, size: usize) -> Result<&'a [u8], DecodeError> {
        if self.len() < size {
            return Err(DecodeError::OutOfData);
        }
        let bz = &self[0..size];
        *self = &self[size..];
        Ok(bz)
    }

    fn is_done(&self) -> Result<(), DecodeError> {
        if !self.is_empty() {
            return Err(DecodeError::InvalidData);
        }
        Ok(())
    }
}