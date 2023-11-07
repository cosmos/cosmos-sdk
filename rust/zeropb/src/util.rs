use crate::error::Error;

pub(crate) const MAX_EXTENT: usize = 0x10000 - 2;

#[inline]
pub(crate) fn resolve_rel_ptr(base: *const u8, offset: i16, min_len: u16) -> *const u8 {
    let buf_start = base as usize & !0xFFFF;
    let target = (base as isize + offset as isize) as usize;
    assert!(target >= buf_start);
    let buf_end = buf_start + 0xFFFF - 2;
    assert!((target + min_len as usize) < buf_end);
    target as *const u8
}

#[inline]
pub(crate) unsafe fn resolve_start_extent(base_ptr: *const u8) -> (usize, *mut u16) {
    let start = (base_ptr as usize) & !0xFFFF;
    (start, (start + MAX_EXTENT) as *mut u16)
}

#[inline]
pub(crate) unsafe fn alloc_rel_ptr(
    base_ptr: *const u8,
    len: usize,
    align: usize,
) -> Result<(i16, *mut ()), Error> {
    let (start, extent_ptr) = resolve_start_extent(base_ptr);
    let alloc_start = (*extent_ptr) as usize;
    // align alloc_start to align
    let alloc_start = (alloc_start + align - 1) & !(align - 1);
    let target = start + alloc_start;
    let base = base_ptr as usize;
    let offset = target - base;
    if offset > i16::MAX as usize {
        return Err(Error::OutOfBounds);
    }

    let next_extent = alloc_start + len;
    if next_extent > MAX_EXTENT {
        return Err(Error::OutOfMemory);
    }

    *extent_ptr = next_extent as u16;
    Ok((offset as i16, target as *mut ()))
}

#[inline]
pub(crate) fn align_addr(addr: usize, align: usize) -> usize {
    (addr + align - 1) & !(align - 1)
}