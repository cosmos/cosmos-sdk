use bump_scope::{BumpScope, BumpBox};
use crate::encoder::EncodeError;

// pub trait WriterFactory {
//     type Writer: Writer;
//     fn new(&self, size: Option<usize>) -> Self::Writer;
// }
//
pub trait Writer {
    fn new(size: Option<usize>) -> Self;
    // type Output;
    fn write(&mut self, bytes: &[u8]) -> Result<(), EncodeError>;
    // fn finish(self) -> Result<Self::Output, EncodeError>;
}

// pub struct BumpWriterFactory<'a> {
//     scope: &'a mut BumpScope<'a>,
// }
//
// impl<'a> BumpWriterFactory<'a> {
//     pub fn new(scope: &'a mut BumpScope<'a>) -> BumpWriterFactory<'a> {
//         BumpWriterFactory {
//             scope,
//         }
//     }
// }
//
//
// pub struct SliceWriter<'a> {
//     buf: &'a mut [u8],
//     pos: usize,
// }
//
// impl<'a> SliceWriter<'a> {
//     pub fn new(buf: &'a mut [u8]) -> SliceWriter<'a> {
//         SliceWriter {
//             buf,
//             pos: 0,
//         }
//     }
//
//     pub fn written(&self) -> usize {
//         self.pos
//     }
// }
//
// impl<'a> Writer for SliceWriter<'a> {
//     type Output = &'a [u8];
//
//     fn write(&mut self, bytes: &[u8]) -> Result<(), EncodeError> {
//         if self.pos + bytes.len() > self.buf.len() {
//             return Err(EncodeError::OutOfSpace);
//         }
//         self.buf[self.pos..self.pos + bytes.len()].copy_from_slice(bytes);
//         self.pos += bytes.len();
//         Ok(())
//     }
//
//     fn finish(self) -> Result<Self::Output, EncodeError> {
//         Ok(&self.buf[0..self.pos])
//     }
// }

#[cfg(feature = "std")]
impl Writer for alloc::vec::Vec<u8> {
    fn new(size: Option<usize>) -> Self {
        match size {
            Some(size) => alloc::vec::Vec::with_capacity(size),
            None => alloc::vec::Vec::new(),
        }
    }

    fn write(&mut self, bytes: &[u8]) -> Result<(), EncodeError> {
        self.extend_from_slice(bytes);
        Ok(())
    }
}

pub trait ReverseWriterFactory {
    type Writer: ReverseWriter;
    fn new(&self, size: usize) -> Self::Writer;
}

pub trait ReverseWriter {
    type Output;
    fn write(&mut self, bytes: &[u8]) -> Result<(), EncodeError>;
    fn pos(&self) -> usize;
    fn finish(self) -> Result<Self::Output, EncodeError>;
}

impl <'a> ReverseWriterFactory for BumpScope<'a> {
    type Writer = ReverseSliceWriter<'a>;

    fn new(&self, size: usize) -> Self::Writer {
        let b = self.alloc_slice_fill(size, 0);
        ReverseSliceWriter {
            buf: b,
            pos: size,
        }
    }
}


pub struct ReverseSliceWriter<'a> {
    buf: BumpBox<'a, [u8]>,
    pos: usize,
}

impl<'a> ReverseWriter for ReverseSliceWriter<'a> {
    type Output = BumpBox<'a, [u8]>;

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

    fn finish(self) -> Result<Self::Output, EncodeError> {
        Ok(self.buf)
    }
}
