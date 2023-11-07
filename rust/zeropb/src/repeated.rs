use core::iter::IntoIterator;
use crate::util::{resolve_rel_ptr, resolve_start_extent, MAX_EXTENT};
use crate::Error;
use core::iter::Iterator;
use core::marker::PhantomData;
use core::mem::{align_of, size_of};
use core::ptr::{NonNull, null};

use crate::zerocopy::ZeroCopy;

#[repr(C)]
pub struct Repeated<T> {
    offset: i16,
    length: u16,
    _phantom: PhantomData<[T]>,
}

unsafe impl<T: ZeroCopy> ZeroCopy for Repeated<T> {}

#[repr(C)]
struct RepeatedSegmentHeader {
    used: u8,
    capacity: u8,
    next_offset: i16,
}

impl<'a, T> Repeated<T> {
    pub fn start_write(&'a mut self) -> Result<RepeatedWriter<'a, T>, Error> {
        unsafe {
            let base = (self as *const Self).cast::<u8>();
            let (buf_start, extent_ptr) = resolve_start_extent(base);
            let cur_extent = *extent_ptr;
            if cur_extent as usize == MAX_EXTENT {
                return Err(Error::OutOfMemory);
            }

            let mut writer = RepeatedWriter {
                ptr: self,
                buf_start,
                extent_ptr,
                cur_segment: core::ptr::null_mut(),
                write_head: core::ptr::null_mut(),
            };

            writer.new_segment()?;

            Ok(writer)
        }
    }

    pub fn len(&self) -> usize {
        self.length as usize
    }
}

impl<'a, T> IntoIterator for &'a Repeated<T> {
    type Item = &'a T;
    type IntoIter = RepeatedIter<'a, T>;

    fn into_iter(self) -> Self::IntoIter {
        let base = NonNull::from(self).as_ptr() as *const u8;
        let (buf_start, _) = unsafe { resolve_start_extent(base) };

        let mut iter = RepeatedIter {
            ptr: self,
            buf_start,
            read_head: null(),
            segment_i: 0,
            cur_segment: null(),
        };

        if self.length > 0 {
            iter.load_next_segment(self.offset as u16);
        }

        iter
    }
}

pub struct RepeatedWriter<'a, T> {
    ptr: &'a mut Repeated<T>,
    buf_start: usize,
    extent_ptr: *mut u16,
    cur_segment: *mut RepeatedSegmentHeader,
    write_head: *mut T,
}

const REPEATED_SEGMENT_HEADER_SIZE: usize = size_of::<RepeatedSegmentHeader>();
const REPEATED_SEGMENT_HEADER_ALIGN: usize = align_of::<RepeatedSegmentHeader>();

impl<'a, T> RepeatedWriter<'a, T> {
    fn new_segment(&mut self) -> Result<(), Error> {
        unsafe {
            let cur_extent = *self.extent_ptr;
            let write_head = (self.buf_start + cur_extent as usize) as *mut u8;

            // align write_head to RepeatedSegmentHeader
            let write_head = write_head.add(REPEATED_SEGMENT_HEADER_ALIGN - (write_head as usize & (REPEATED_SEGMENT_HEADER_ALIGN - 1)));
            if self.cur_segment != core::ptr::null_mut() {
                (*self.cur_segment).next_offset = (write_head as usize - self.buf_start) as i16;
            }
            let cur_segment = write_head as *mut RepeatedSegmentHeader;

            // advance write_head and align to T
            let write_head = write_head.add(REPEATED_SEGMENT_HEADER_SIZE);
            let align_t: usize = align_of::<T>();
            let write_head = write_head.add(align_t - (write_head as usize & (align_t - 1)));

            // set capacity to the number of Ts that can fit in 64 bytes or 1, whichever is greater
            let size_t: usize = size_of::<T>();
            let capacity = core::cmp::max(64 / size_t, 1) as u8;
            (*cur_segment).capacity = capacity;

            // update extent and check bounds
            let write_limit = write_head.add(capacity as usize);
            let next_extent = write_limit as usize - self.buf_start;
            if next_extent > MAX_EXTENT {
                return Err(Error::OutOfMemory);
            }
            *self.extent_ptr = next_extent as u16;

            self.cur_segment = cur_segment;
            self.write_head = write_head as *mut T;

            Ok(())
        }
    }

    pub fn append(&mut self) -> Result<&mut T, Error> {
        unsafe {
            let cur_segment = &mut *self.cur_segment;
            if cur_segment.used == cur_segment.capacity {
                self.new_segment()?;
            }

            let ret = &mut *self.write_head;
            self.write_head = self.write_head.add(1);
            cur_segment.used += 1;
            self.ptr.length += 1;
            return Ok(ret);
        }
    }
}

pub struct RepeatedIter<'a, T> {
    ptr: &'a Repeated<T>,
    buf_start: usize,
    read_head: *const T,
    segment_i: u8,
    cur_segment: *const RepeatedSegmentHeader,
}

impl<'a, T> RepeatedIter<'a, T> {
    fn load_next_segment(&mut self, offset: u16) {
        unsafe {
            let read_head = resolve_rel_ptr(self.buf_start as *const u8, offset as i16, 0) as *const u8;
            // align to RepeatedSegmentHeader
            let read_head = read_head.add(REPEATED_SEGMENT_HEADER_ALIGN - (read_head as usize & (REPEATED_SEGMENT_HEADER_ALIGN - 1)));
            self.cur_segment = read_head as *const RepeatedSegmentHeader;
            // check buffer overflow
            let read_head = read_head.add(REPEATED_SEGMENT_HEADER_SIZE);
            assert!(read_head as usize <= self.buf_start + MAX_EXTENT);

            self.segment_i = 0;

            // align read head to T
            let align_t = align_of::<T>();
            let read_head = read_head.add(align_t - (read_head as usize & (align_t - 1)));
            self.read_head = read_head as *const T;

            // check buffer overflow
            let read_limit = self.read_head.add((*self.cur_segment).capacity as usize);
            assert!(read_limit as usize <= self.buf_start + MAX_EXTENT);
        }
    }
}


impl<'a, T> Iterator for RepeatedIter<'a, T> {
    type Item = &'a T;

    fn next(&mut self) -> Option<Self::Item> {
        unsafe {
            if self.cur_segment == null() {
                return None;
            }

            let cur_segment = &*self.cur_segment;
            if self.segment_i == cur_segment.used {
                // resolve next segment
                let next_offset = cur_segment.next_offset;
                if next_offset == 0 {
                    self.cur_segment = null();
                    return None;
                }

                self.load_next_segment(next_offset as u16);
            }

            let ret = &*self.read_head;
            self.read_head = self.read_head.add(size_of::<T>());
            self.segment_i += 1;
            Some(ret)
        }
    }
}

pub struct ScalarRepeated<T> {
    repeated: Repeated<T>,
}

unsafe impl<T: ZeroCopy + Copy> ZeroCopy for ScalarRepeated<T> {}

impl<T> ScalarRepeated<T> {
    pub fn start_write(&mut self) -> Result<ScalarRepeatedWriter<T>, Error> {
        self.repeated.start_write().map(|writer| ScalarRepeatedWriter { writer })
    }
}

impl<'a, T> IntoIterator for &'a ScalarRepeated<T> {
    type Item = &'a T;
    type IntoIter = RepeatedIter<'a, T>;

    fn into_iter(self) -> Self::IntoIter {
        self.repeated.into_iter()
    }
}

pub struct ScalarRepeatedWriter<'a, T> {
    writer: RepeatedWriter<'a, T>,
}

impl<'a, T> ScalarRepeatedWriter<'a, T> {
    pub fn append(&mut self, value: T) -> Result<(), Error> {
        unsafe {
            let ptr = self.writer.append()?;
            *ptr = value;
            Ok(())
        }
    }
}

#[cfg(test)]
mod tests {
    use crate::repeated::Repeated;
    use crate::{Root, ZeroCopy};

    #[repr(C)]
    struct A {
        repeated: Repeated<B>,
    }

    unsafe impl ZeroCopy for A {}

    #[repr(C)]
    struct B {
        a: u32,
        b: u16,
    }

    unsafe impl ZeroCopy for B {}

    #[test]
    fn repeated() {
        let mut a = Root::<A>::new();
        let mut writer = a.repeated.start_write().unwrap();
        let mut b = writer.append().unwrap();
        b.a = 1;
        b.b = 2;
        b = writer.append().unwrap();
        b.a = 3;
        b.b = 4;
        b = writer.append().unwrap();
        b.a = 5;
        b.b = 6;
        let mut iter = a.repeated.into_iter();
        let b = iter.next().unwrap();
        assert_eq!(b.a, 1);
        assert_eq!(b.b, 2);
        let b = iter.next().unwrap();
        assert_eq!(b.a, 3);
        assert_eq!(b.b, 4);
        let b = iter.next().unwrap();
        assert_eq!(b.a, 5);
        assert_eq!(b.b, 6);
        assert!(iter.next().is_none());
    }
}
