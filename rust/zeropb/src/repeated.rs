use core::iter::Iterator;
use core::marker::PhantomData;
use core::ptr::NonNull;
use crate::Error;
use crate::util::{MAX_EXTENT, resolve_rel_ptr, resolve_start_extent};

use crate::zerocopy::ZeroCopy;

#[repr(C)]
pub struct Repeated<T> {
    offset: i16,
    length: u16,
    _phantom: PhantomData<[T]>,
}

unsafe impl<T: ZeroCopy> ZeroCopy for Repeated<T> {}

impl<'a, T> Repeated<T> {
    pub fn start_write(&'a mut self) -> Result<RepeatedWriter<'a, T>, Error> {
        unsafe {
            let base = (self as *const Self).cast::<u8>();
            let (start, extent_ptr) = resolve_start_extent(base);
            let last_extent = *extent_ptr;
            if last_extent as usize == MAX_EXTENT {
                return Err(Error::OutOfMemory);
            }

            let write_head = (start + last_extent as usize) as *mut u8;

            // align write head to T
            let align_to = core::mem::align_of::<T>();
            let write_head = (write_head as usize + align_to - 1) & !(align_to - 1);
            self.offset = (write_head - base as usize) as i16;
            self.length = 0;

            Ok(RepeatedWriter {
                ptr: self,
                extent_ptr,
                write_head: write_head as *mut T,
                last_extent,
                size_t: core::mem::size_of::<T>(),
            })
        }
    }
}

impl<'a, T> IntoIterator for &'a Repeated<T> {
    type Item = &'a T;
    type IntoIter = RepeatedIter<'a, T>;

    fn into_iter(self) -> Self::IntoIter {
        let mut iter = RepeatedIter {
            ptr: self,
            read_head: core::ptr::null(),
            i: 0,
            size_t: core::mem::size_of::<T>(),
        };

        if self.offset == 0 {
            return iter;
        }

        let base = NonNull::from(self).as_ptr() as *const u8;
        let target = resolve_rel_ptr(base, self.offset, self.length);
        iter.read_head = target as *const T;
        return iter;
    }
}

pub struct RepeatedWriter<'a, T> {
    ptr: &'a mut Repeated<T>,
    extent_ptr: *mut u16,
    write_head: *mut T,
    last_extent: u16,
    size_t: usize,
}

impl<'a, T> RepeatedWriter<'a, T> {
    pub fn append(&mut self) -> Result<&mut T, Error> {
        unsafe {
            let extent = *self.extent_ptr;
            if extent != self.last_extent {
                return Err(Error::InvalidState);
            }

            self.ptr.length += 1;
            let next_extent = extent as usize + self.size_t;
            if next_extent > MAX_EXTENT {
                return Err(Error::OutOfMemory);
            }

            let target = self.write_head;
            self.write_head = self.write_head.add(self.size_t);
            self.last_extent = next_extent as u16;
            *self.extent_ptr = next_extent as u16;

            Ok(&mut *target)
        }
    }
}

pub struct RepeatedIter<'a, T> {
    ptr: &'a Repeated<T>,
    read_head: *const T,
    size_t: usize,
    i: u16,
}

impl<'a, T> Iterator for RepeatedIter<'a, T> {
    type Item = &'a T;

    fn next(&mut self) -> Option<Self::Item> {
        if self.i == self.ptr.length || self.read_head == core::ptr::null() {
            return None;
        }

        unsafe {
            let ret = &*self.read_head;
            self.read_head = self.read_head.add(self.size_t);
            self.i += 1;
            Some(ret)
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
