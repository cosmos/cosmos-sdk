use std::alloc::{alloc_zeroed, Layout};
use std::marker::{PhantomData, PhantomPinned};
use std::ops::Deref;
use core::str::{from_utf8_unchecked};

struct Buffer<T> {
    data: *mut u8,
    size: usize,
    capacity: usize,
    _phantom: PhantomData<T>,
}

impl <T> Buffer<T> {
    fn new() -> Self {
        const DEFAULT_CAP: usize = 2048;
        Self {
            data: unsafe { alloc_zeroed(Layout::from_size_align_unchecked(DEFAULT_CAP, 1)) },
            size: 0,
            capacity: DEFAULT_CAP,
            _phantom: PhantomData,
        }
    }

}

trait Buf {
    fn alloc(&mut self, size: usize) -> *mut u8;
}

impl<T> Buf for Buffer<T> {
    fn alloc(&mut self, size: usize) -> *mut u8 {
        let new_size = self.size + size;
        if new_size > self.capacity {
            let new_capacity = self.capacity * 2;
            let new_data = unsafe { alloc_zeroed(Layout::from_size_align_unchecked(new_capacity, 1)) };
            unsafe {
                std::ptr::copy_nonoverlapping(self.data, new_data, self.size);
            }
            self.data = new_data;
            self.capacity = new_capacity;
        }
        let ptr = unsafe { self.data.add(self.size) };
        self.size = new_size;
        ptr
    }
}

struct Bytes {
    arr: Array<u8>,
}

impl <'a> Deref for Bytes {
    type Target = Option<&'a [u8]>;

    fn deref(&self) -> &Self::Target {
        self.arr.deref()
    }
}

struct String {
    arr: Bytes
}

impl <'a> Deref for String {
    type Target = Option<&'a str>;

    fn deref(&'a self) -> &Self::Target {
        &self.arr.deref().map(|arr| unsafe { from_utf8_unchecked(arr) })
    }
}

impl String {
    fn set(&mut self, buffer: &mut dyn Buf, value: &str) {
        self.arr.set(buffer, value.as_bytes())
    }
}

impl Bytes {
    fn set(&mut self, buffer: &mut dyn Buf, value: &[u8]) {
        let mut mem = buffer.alloc(value.len());
        unsafe {
            std::ptr::copy_nonoverlapping(value.as_ptr(), mem, value.len());
        }
        self.rel_ptr.set(mem)
    }
}

#[repr(C)]
struct RawRelPtr {
    offset: u16,
    _phantom: PhantomPinned,
}

impl RawRelPtr {
    fn get(&self) -> *u8 {
        (self as *u8) + self.offset
    }

    fn set(&mut self, value: *mut u8) {
        self.offset = value - self as *mut u8
    }
}

#[repr(C)]
struct Array<T> {
    ptr: RawRelPtr,
}

impl <'a, T> Deref for Array<T> {
    type Target = Option<&'a [T]>;

    fn deref(&'a self) -> &Self::Target {
        if self.rel_ptr.offset == 0 {
            &None
        } else {
            let ptr = self.rel_ptr.get();
            let len = unsafe{*(ptr as *u16)};
            &Some(unsafe { std::slice::from_raw_parts(ptr + 2, len as usize) })
        }
    }
}

#[repr(C)]
struct RelPtr<T: ?Sized> {
    offset: u16,
    _phantom: PhantomData<T>,
}

struct Enum<T, const MaxValue: u8> {
    value: u8
}

impl <T, const MaxValue: u8> Enum<T, MaxValue> {
    fn get(&self) -> Result<&T, u8> {
        if self.value > MaxValue {
            Err(self.value)
        } else {
            Ok(&self.value as T)
        }
    }

    fn set(&mut self, value: T) {
        self.value = value
    }
}

struct OneOf<T, const MaxValue: u8> {
    value: T
}

impl <T, const MaxValue: u8> OneOf<T, MaxValue> {
    fn get(&self) -> Result<&T, u8> {
        let discriminant = unsafe { *<*const _>::from(self).cast::<u8>() };
        if discriminant > MaxValue {
            Err(discriminant)
        } else {
            Ok(&self.value)
        }
    }

    fn set(&mut self, value: T) {
        self.value = value
    }
}
