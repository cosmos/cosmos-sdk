extern crate core;
use core::borrow::Borrow;
use core::ops::{Deref, DerefMut, Drop};
use core::ptr::null;
use core::default::Default;
use core::marker::Sized;

pub struct RawBytes {
    bytes: *const u8,
    len: usize,
    free: fn(*mut u8, usize),
}

impl Default for RawBytes {
    fn default() -> Self {
        RawBytes {
            bytes: null(),
            len: 0,
            free: |_, _| {},
        }
    }
}

impl Drop for RawBytes {
    fn drop(&mut self) {
        (self.free)(self.bytes as *mut u8, self.len);
    }
}

pub struct RawString(RawBytes);

impl Default for RawString {
    fn default() -> Self {
        RawString(RawBytes::default())
    }
}

pub struct RawBox<T: ?Sized> {
    t: *mut T,
    len: usize,
    free: fn(*mut u8, usize),
}

impl<T: ?Sized> Drop for RawBox<T> {
    fn drop(&mut self) {
        (self.free)(self.t as *mut u8, self.len);
    }
}

impl<T: ?Sized> Deref for RawBox<T> {
    type Target = T;

    fn deref(&self) -> &Self::Target {
        unsafe { &*self.t }
    }
}

impl<T: ?Sized> Borrow<T> for RawBox<T> {
    fn borrow(&self) -> &T {
        unsafe { &*self.t }
    }
}
