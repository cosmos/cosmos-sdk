#![feature(offset_of)]

extern crate alloc;
extern crate core;

use core::{
    marker::{PhantomData},
    ops::{Deref, DerefMut},
    slice::from_raw_parts,
    str::from_utf8_unchecked,
    borrow::{Borrow, BorrowMut},
    mem::{size_of},
    ptr,
};
use alloc::alloc::{alloc_zeroed, Layout};
use thiserror::Error;

pub trait ZeroCopy {}

impl ZeroCopy for bool {}

impl ZeroCopy for rend::i32_le {}

impl ZeroCopy for rend::u32_le {}

impl ZeroCopy for rend::i64_le {}

impl ZeroCopy for rend::u64_le {}

pub struct Root<T: ZeroCopy> {
    buf: *mut u8,
    _phantom: PhantomData<T>,
}

const MAX_EXTENT: usize = 0x10000 - 2;

impl<T: ZeroCopy> Root<T> {
    fn new() -> Self {
        unsafe {
            let buf = alloc_zeroed(Layout::from_size_align_unchecked(0x10000, 0x10000));
            assert!(!buf.is_null());
            assert_eq!((buf as usize) & 0xFFFF, 0);
            let extent_ptr = buf.offset(MAX_EXTENT as isize) as *mut u16;
            let size_of_t = size_of::<T>();
            assert!(size_of_t <= MAX_EXTENT);
            *extent_ptr = size_of_t as u16;
            Self {
                buf,
                _phantom: PhantomData,
            }
        }
    }
}

impl<T: ZeroCopy> Borrow<T> for Root<T>  {
    fn borrow(&self) -> &T {
        unsafe {
            &*self.buf.cast::<T>()
        }
    }
}

impl<T: ZeroCopy> BorrowMut<T> for Root<T>  {
    fn borrow_mut(&mut self) -> &mut T {
        unsafe {
            &mut *self.buf.cast::<T>()
        }
    }
}

impl<T: ZeroCopy> Deref for Root<T>  {
    type Target = T;

    fn deref(&self) -> &Self::Target {
        unsafe {
            &*self.buf.cast::<T>()
        }
    }
}

impl<T: ZeroCopy> DerefMut for Root<T>  {
    fn deref_mut(&mut self) -> &mut Self::Target {
        unsafe {
            &mut *self.buf.cast::<T>()
        }
    }
}

#[inline]
fn resolve_rel_ptr(base: *const u8, offset: i16, min_len: u16) -> usize {
    let buf_start = base as usize & !0xFFFF;
    let target = (base as isize + offset as isize) as usize;
    assert!(target >= buf_start);
    let buf_end = buf_start + 0xFFFF - 2;
    assert!((target + min_len as usize) < buf_end);
    target
}

#[inline]
unsafe fn alloc_rel_ptr(base_ptr: *const u8, len: usize, align: usize) -> Result<(i16, *mut ()), Error> {
    let base = base_ptr as usize;
    let start = base & !0xFFFF;
    let end = start + MAX_EXTENT;
    let extent_ptr = end as *mut u16;
    let alloc_start = (*extent_ptr) as usize;
    // align alloc_start to align
    let alloc_start = (alloc_start + align - 1) & !(align - 1);
    let next_extent = alloc_start + len;
    if next_extent > MAX_EXTENT {
        return Err(Error::OutOfMemory);
    }

    *extent_ptr = next_extent as u16;
    let target = start + alloc_start;
    Ok(((target - base) as i16, target as *mut ()))
}


#[repr(C)]
pub struct Bytes {
    offset: i16,
    length: u16,
    _phantom: PhantomData<[u8]>,
}

#[derive(Error, Debug)]
enum Error {
    #[error("out of memory")]
    OutOfMemory
}

impl ZeroCopy for Bytes {}

impl Bytes {
    fn set(&mut self, content: &[u8]) -> Result<(), Error> {
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
}

impl<'a> Borrow<[u8]> for Bytes {
    fn borrow(&self) -> &[u8] {
        unsafe {
            let base = (self as *const Self).cast::<u8>();
            let target = resolve_rel_ptr(base , self.offset, self.length);
            from_raw_parts(target as *const u8, self.length as usize)
        }
    }
}

#[repr(C)]
pub struct Str {
    ptr: Bytes,
    _phantom: PhantomData<str>,
}

impl ZeroCopy for Str {}

impl Str {
    fn set(&mut self, content: &str) -> Result<(), Error> {
        self.ptr.set(content.as_bytes())
    }
}

impl<'a> Borrow<str> for Str {
    fn borrow(&self) -> &str {
        unsafe {
            unsafe {
                from_utf8_unchecked(self.ptr.borrow())
            }
        }
    }
}

#[repr(C)]
pub struct RepeatedPtr<T> {
    offset: i16,
    length: u16,
    _phantom: PhantomData<[T]>,
}

struct RepeatedSegmentHeader {
    capacity: u16,
    next_offset: i16,
}

// #[repr(C)]
// pub struct Enum<T: Copy, const MaxValue: u8> {
//     value: u8,
//     _phantom: PhantomData<T>,
// }
//
// impl<T: Copy, const MaxValue: u8> Enum<T, MaxValue> {
//     fn get(&self) -> Result<T, u8> {
//         if self.value > MaxValue {
//             Err(self.value)
//         } else {
//             Ok(self.value as T)
//         }
//     }
//
//     fn set(&mut self, value: T) {
//         self.value = value
//     }
// }
//
// #[repr(C)]
// pub struct OneOf<T, const MaxValue: u32> {
//     value: T
// }
//
// impl <T, const MaxValue: u32> OneOf<T, MaxValue> {
//     fn get(&self) -> Result<&T, u32> {
//         let discriminant = unsafe { *<*const _>::from(self).cast::<u32>() };
//         if discriminant > MaxValue {
//             Err(discriminant)
//         } else {
//             Ok(&self.value)
//         }
//     }
//
//     fn set(&mut self, value: T) {
//         self.value = value
//     }
// }
//
// #[repr(C)]
// pub struct Option<T> {
//     some: bool,
//     value: T,
// }

#[cfg(test)]
mod tests {
    use std::borrow::Borrow;
use crate::{Root, Str, ZeroCopy};

    #[repr(C, align(4))]
    struct A(u8);

    #[repr(C, u8, align(8))]
    enum TestEnum {
        A(A),
        // B(bool),
        // C(u32),
    }

    #[repr(C)]
    struct TestStruct {
        s: Str
    }

    impl ZeroCopy for TestStruct {}

    #[test]
    fn test1() {
        let mut r = Root::<TestStruct>::new();
        r.s.set("hello").unwrap();
        assert_eq!(<Str as Borrow<str>>::borrow(&r.s), "hello");
    }

    #[test]
    fn size() {
        // assert_eq!(std::mem::size_of::<TestEnum>(), 8);
        // assert_eq!(std::mem::offset_of!(TestEnum, A.0), 4);
        // assert_eq!(std::mem::offset_of!(TestEnum, B.0), 4);
        // assert_eq!(std::mem::offset_of!(TestEnum, C.0), 4);

        // let t1 = &mut Test1 {
        //     bytes: BytesPtr {
        //         offset: 0,
        //         length: 0,
        //         _phantom: Default::default(),
        //     }
        // };
        // let t2 = &mut Test1 {
        //     bytes: BytesPtr {
        //         offset: 0,
        //         length: 0,
        //         _phantom: Default::default(),
        //     }
        // };
        // t1.bytes = t2.bytes;
    }
}
