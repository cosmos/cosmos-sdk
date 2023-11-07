use core::iter::IntoIterator;
use crate::rel_ptr::{resolve_rel_ptr, resolve_start_extent, MAX_EXTENT, align_addr};
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

            writer.new_segment(base)?;

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
            iter.load_next_segment(base, self.offset);
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
    fn new_segment(&mut self, base: *const u8) -> Result<(), Error> {
        unsafe {
            let cur_extent = *self.extent_ptr;
            let write_head = self.buf_start + cur_extent as usize;

            // align write_head to RepeatedSegmentHeader
            let write_head = align_addr(write_head, REPEATED_SEGMENT_HEADER_ALIGN);
            // update pointer
            if self.cur_segment == core::ptr::null_mut() {
                self.ptr.offset = (write_head - base as usize) as i16;
            } else {
                (*self.cur_segment).next_offset = (write_head - base as usize) as i16;
            }
            let cur_segment = write_head as *mut RepeatedSegmentHeader;

            // advance write_head and align to T
            let write_head = write_head + REPEATED_SEGMENT_HEADER_SIZE;
            let align_t: usize = align_of::<T>();
            let write_head = align_addr(write_head, align_t);

            // set capacity to the number of Ts that can fit in 32 bytes or 1, whichever is greater
            // 32 bytes is arbitrary and this could be tweaked depending on real world usage
            let size_t: usize = size_of::<T>();
            let capacity = core::cmp::max(32 / size_t, 1) as u8;
            (*cur_segment).capacity = capacity;

            // update extent and check bounds
            let write_limit = write_head + capacity as usize * size_t;
            let next_extent = write_limit - self.buf_start;
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
            let mut cur_segment = &mut *self.cur_segment;
            if cur_segment.used == cur_segment.capacity {
                self.new_segment(self.cur_segment as *const u8)?;
                cur_segment = &mut *self.cur_segment;
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
    fn load_next_segment(&mut self, base: *const u8, offset: i16) {
        unsafe {
            let read_head = resolve_rel_ptr(base, offset, 0) as usize;
            // align to RepeatedSegmentHeader
            let read_head = align_addr(read_head, REPEATED_SEGMENT_HEADER_ALIGN);
            self.cur_segment = read_head as *const RepeatedSegmentHeader;
            // check buffer overflow
            let read_head = read_head + REPEATED_SEGMENT_HEADER_SIZE;
            assert!(read_head <= self.buf_start + MAX_EXTENT);

            self.segment_i = 0;

            // align read head to T
            let align_t = align_of::<T>();
            let read_head = align_addr(read_head, align_t);
            self.read_head = read_head as *const T;

            // check buffer overflow
            let read_limit = read_head + (*self.cur_segment).used as usize * size_of::<T>();
            assert!(read_limit <= self.buf_start + MAX_EXTENT);
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

                self.load_next_segment(self.cur_segment as *const u8, next_offset);
            }

            let ret = &*self.read_head;
            self.read_head = self.read_head.add(1);
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
        let ptr = self.writer.append()?;
        *ptr = value;
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use crate::repeated::{Repeated, ScalarRepeated};
    use crate::{Root, Str, ZeroCopy};

    #[repr(C)]
    struct A {
        repeated: Repeated<B>,
        repeated_u64: ScalarRepeated<u64>,
    }

    unsafe impl ZeroCopy for A {}

    #[repr(C)]
    struct B {
        a: u32,
        b: Str,
        c: u64,
    }

    unsafe impl ZeroCopy for B {}

    #[test]
    fn repeated() {
        // set data
        let mut a = Root::<A>::new();
        let mut writer = a.repeated.start_write().unwrap();
        let mut b = writer.append().unwrap();
        b.a = 1;
        b.c = 2;
        b.b.set("hello").unwrap();
        b = writer.append().unwrap();
        b.a = 3;
        b.c = 4;
        b.b.set("world").unwrap();
        b = writer.append().unwrap();
        b.a = 5;
        b.c = 6;
        b.b.set("foo").unwrap();

        let mut writer = a.repeated_u64.start_write().unwrap();
        for i in 0..100 {
            writer.append(i).unwrap();
        }

        // check data
        let mut iter = a.repeated.into_iter();
        let b = iter.next().unwrap();
        assert_eq!(b.a, 1);
        assert_eq!(b.c, 2);
        assert_eq!(<Str as core::borrow::Borrow<str>>::borrow(&b.b), "hello");
        let b = iter.next().unwrap();
        assert_eq!(b.a, 3);
        assert_eq!(b.c, 4);
        assert_eq!(<Str as core::borrow::Borrow<str>>::borrow(&b.b), "world");
        let b = iter.next().unwrap();
        assert_eq!(b.a, 5);
        assert_eq!(b.c, 6);
        assert_eq!(<Str as core::borrow::Borrow<str>>::borrow(&b.b), "foo");
        assert!(iter.next().is_none());

        let mut iter = a.repeated_u64.into_iter();
        for i in 0..100 {
            assert_eq!(*iter.next().unwrap(), i as u64);
        }
    }
}
