use core::iter::Iterator;
use core::marker::PhantomData;

use crate::zerocopy::ZeroCopy;

#[repr(C)]
pub struct RepeatedPtr<T: ZeroCopy> {
    offset: i16,
    length: u16,
    _phantom: PhantomData<[T]>,
}

unsafe impl <T: ZeroCopy> ZeroCopy for RepeatedPtr<T> {}

pub struct RepeatedPtrIter<'a, T: ZeroCopy> {
    ptr: &'a RepeatedPtr<T>,
    cur_header: *const RepeatedSegmentHeader,
    cur_data: *const T,
    cur_seg_pos: u16,
    cur_pos: u16,
}

impl <'a, T: ZeroCopy> Iterator for RepeatedPtrIter<'a, T> {
    type Item = &'a T;

    fn next(&mut self) -> Option<Self::Item> {
        // if self.cur_header.is_null() {
        //     if self.ptr.offset == 0 {
        //         return None;
        //     }
        //
        //     unsafe {
        //         let base = (self.ptr as *const Self).cast::<u8>();
        //         let target = resolve_rel_ptr(base, self.ptr.offset, self.ptr.length) as *mut u8;
        //         self.cur_header = target as *mut RepeatedSegmentHeader;
        //         self.cur_data = target.add(align_of::<T>()) as *mut T;
        //         self.cur_pos = 0;
        //     }
        // }

        todo!()
    }
}

// #[repr(C)]
// pub struct Repeated<'a, T: ZeroCopy> {
//     ptr: &'a mut RepeatedPtr<T>,
//     cur_header: *mut RepeatedSegmentHeader,
//     cur_data: *mut T,
// }

struct RepeatedSegmentHeader {
    capacity: u16,
    next_offset: i16,
}

#[cfg(test)]
mod tests {
    use std::alloc;

    #[repr(C)]
    struct TestStruct {
        a: u32,
        b: u16,
    }

    #[test]
    fn bad_align() {
        unsafe {
            let ptr = alloc::alloc_zeroed(alloc::Layout::from_size_align_unchecked(128, 8));
            let ps = ptr.add(8) as *mut TestStruct;
            (*ps).a = 0x12345678;
        }
    }
}
