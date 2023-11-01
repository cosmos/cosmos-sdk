extern crate alloc;
extern crate core;

use core::{
    marker::{PhantomData, PhantomPinned},
    ops::{Deref, DerefMut},
    slice::from_raw_parts,
    str::from_utf8_unchecked,
    ptr,
};
use alloc::alloc::{alloc_zeroed, Layout};

pub struct Root<T> {
    buf: *mut u8,
    _phantom: PhantomData<T>,
}

impl<T> Root<T> {
    fn new() -> Self {
        Self {
            buf: unsafe { alloc_zeroed(Layout::from_size_align_unchecked(0xFFFF, 0xFFFF)) },
            _phantom: PhantomData,
        }
    }
}

#[repr(C)]
struct RawRelPtr {
    offset: i16,
    _phantom: PhantomPinned,
}

fn resolve_rel_ptr(base: usize, offset: i16, min_len: u16) -> usize {
    let buf_start = base as usize & !0xFFFF;
    let target = (base as isize + offset as isize) as usize;
    assert!(target >= buf_start);
    let buf_end = buf_start + 0xFFFF;
    assert!((target + min_len) < buf_end);
    target
}

impl RawRelPtr {
    /// Attempts to create a new `RawRelPtr` in-place between the given `from` and `to` positions.
    ///
    /// # Safety
    ///
    /// - `out` must be located at position `from`
    /// - `to` must be a position within the archive
    // #[inline]
    // pub unsafe fn try_emplace(from: usize, to: usize, out: *mut Self) -> Result<(), OffsetError> {
    //     let offset = O::between(from, to)?;
    //     ptr::addr_of_mut!((*out).offset).write(offset);
    //     Ok(())
    // }

    /// Creates a new `RawRelPtr` in-place between the given `from` and `to` positions.
    ///
    /// # Safety
    ///
    /// - `out` must be located at position `from`
    /// - `to` must be a position within the archive
    /// - The offset between `from` and `to` must fit in an `isize` and not exceed the offset
    ///   storage
    #[inline]
    pub unsafe fn emplace(from: usize, to: usize, out: *mut Self) {
        Self::try_emplace(from, to, out).unwrap();
    }

    /// Gets the base pointer for the relative pointer.
    #[inline]
    pub fn base(&self) -> (*const u8, usize) {
        let base = (self as *const Self).cast::<u8>();
        let base_start = base as usize & !0xFFFF;
        (base, base_start)
    }

    /// Gets the mutable base pointer for the relative pointer.
    #[inline]
    pub fn base_mut(&mut self) -> (*mut u8, usize) {
        let base = (self as *mut Self).cast::<u8>();
        let base_start = base as usize & !0xFFFF;
        (base, base_start)
    }

    /// Gets the offset of the relative pointer from its base.
    #[inline]
    pub fn offset(&self) -> isize {
        self.offset.to_isize()
    }

    /// Gets whether the offset of the relative pointer is 0.
    #[inline]
    pub fn is_null(&self) -> bool {
        self.offset() == 0
    }

    /// Calculates the memory address being pointed to by this relative pointer.
    #[inline]
    pub fn as_ptr(&self) -> *const () {
        unsafe {
            let (base, buf_start) = self.base();
            let offset = self.offset();
            let target = base.offset(offset).cast();
            let buf_end = buf_start + 0xFFFF;
            assert!((target as usize) >= buf_start);
            assert!((target as usize) < buf_end);
            target
        }
    }

    /// Returns an unsafe mutable pointer to the memory address being pointed to
    /// by this relative pointer.
    #[inline]
    pub fn as_mut_ptr(&mut self) -> *mut () {
        unsafe { self.base_mut().offset(self.offset()).cast() }
    }
}

#[repr(C)]
struct RelLenPtr {
    _phantom: PhantomPinned,
}

#[repr(C)]
pub struct BytesPtr {
    offset: i16,
    length: u16,
    _phantom: PhantomData<[u8]>,
}

impl <'a> Deref for BytesPtr {
    type Target = [u8];

    fn deref(&self) -> &Self::Target {
        unsafe {
            unsafe {
                let base = (self as *const Self).cast::<u8>();
                let target = resolve_rel_ptr(base as usize, self.offset, self.length);
                from_raw_parts(target as *const u8, self.length as usize)
            }
        }
    }
}

#[repr(C)]
pub struct StringPtr {
    ptr: BytesPtr,
    _phantom: PhantomData<str>,
}

impl <'a> Deref for StringPtr {
    type Target = str;

    fn deref(&self) -> &Self::Target {
        unsafe {
            unsafe {
                from_utf8_unchecked(self.ptr.deref())
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

struct RepeatedSegment<T> {

}

// #[repr(C)]
// pub struct Enum<T, const MaxValue: u32> {
//     value: u32,
//     _phantom: PhantomData<T>,
// }
//
// impl <T, const MaxValue: u32> Enum<T, MaxValue> {
//     fn get(&self) -> Result<&T, u32> {
//         if self.value > MaxValue {
//             Err(self.value)
//         } else {
//             Ok(&self.value as T)
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
    use crate::Root;

    #[test]
    fn test_root() {
        let r = Root::<()>::new();
    }
}
