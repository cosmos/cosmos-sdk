use core::marker::PhantomData;

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

