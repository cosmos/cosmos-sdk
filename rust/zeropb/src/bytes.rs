use core::{borrow::Borrow, marker::PhantomData, ptr, slice::from_raw_parts};

use crate::error::Error;
use crate::util::{alloc_rel_ptr, resolve_rel_ptr, resolve_start_extent, MAX_EXTENT};
use crate::zerocopy::ZeroCopy;

#[repr(C)]
pub struct Bytes {
    offset: i16,
    length: u16,
    _phantom: PhantomData<[u8]>,
}

unsafe impl ZeroCopy for Bytes {}

impl Bytes {
    pub fn set(&mut self, content: &[u8]) -> Result<(), Error> {
        unsafe {
            let base = (self as *const Self).cast::<u8>();
            let len = content.len();
            let (offset, target) = alloc_rel_ptr(base, len, 1)?;
            self.offset = offset;
            self.length = len as u16;
            ptr::copy_nonoverlapping(content.as_ptr(), target as *mut u8, len);
            Ok(())
        }
    }

    pub fn new_writer(&mut self) -> Result<BytesWriter, Error> {
        unsafe {
            let base = (self as *const Self).cast::<u8>();
            let (start, extent_ptr) = resolve_start_extent(base);
            let last_extent = *extent_ptr;
            if last_extent as usize == MAX_EXTENT {
                return Err(Error::OutOfMemory);
            }

            let write_head = (start + last_extent as usize) as *mut u8;
            self.offset = (write_head as usize - base as usize) as i16;
            self.length = 0;

            Ok(BytesWriter {
                bz: self,
                extent_ptr,
                write_head,
                last_extent,
            })
        }
    }
}

impl<'a> Borrow<[u8]> for Bytes {
    fn borrow(&self) -> &[u8] {
        unsafe {
            let base = (self as *const Self).cast::<u8>();
            let target = resolve_rel_ptr(base, self.offset, self.length);
            from_raw_parts(target, self.length as usize)
        }
    }
}

pub struct BytesWriter<'a> {
    bz: &'a mut Bytes,
    extent_ptr: *mut u16,
    write_head: *mut u8,
    last_extent: u16,
}

impl<'a> BytesWriter<'a> {
    pub fn write(&mut self, content: &[u8]) -> Result<(), Error> {
        unsafe {
            let extent = *self.extent_ptr;
            if extent != self.last_extent {
                return Err(Error::InvalidState);
            }

            let len = content.len();
            self.bz.length += len as u16;
            let next_extent = extent as usize + len;
            if next_extent > crate::util::MAX_EXTENT {
                return Err(Error::OutOfMemory);
            }

            ptr::copy_nonoverlapping(content.as_ptr(), self.write_head, len);
            self.write_head = self.write_head.add(len);
            self.last_extent = next_extent as u16;
            *self.extent_ptr = next_extent as u16;

            Ok(())
        }
    }
}
