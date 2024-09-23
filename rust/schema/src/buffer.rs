use bump_scope::BumpScope;
use crate::encoder::EncodeError;

pub trait Writer {
    fn write(&mut self, bytes: &[u8]) -> Result<(), EncodeError>;
}

pub struct SliceWriter<'a> {
    buf: &'a mut [u8],
    pos: usize,
}

impl<'a> SliceWriter<'a> {
    pub fn new(buf: &'a mut [u8]) -> SliceWriter<'a> {
        SliceWriter {
            buf,
            pos: 0,
        }
    }

    pub fn written(&self) -> usize {
        self.pos
    }
}

impl<'a> Writer for SliceWriter<'a> {
    fn write(&mut self, bytes: &[u8]) -> Result<(), EncodeError> {
        if self.pos + bytes.len() > self.buf.len() {
            return Err(EncodeError::OutOfSpace);
        }
        self.buf[self.pos..self.pos + bytes.len()].copy_from_slice(bytes);
        self.pos += bytes.len();
        Ok(())
    }
}

