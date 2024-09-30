use core::cell::Cell;
use core::ptr::NonNull;


// Work in progress on a custom bump allocator to avoid a dependency and customize chunk size.
struct Bump {
    cur: *mut Chunk
}

struct Chunk {
    pos: Cell<*mut u8>,
    end: *mut u8,
    prev: Option<*mut u8>,
    next: Cell<Option<*mut u8>>
}

impl Bump {
    unsafe fn allocate(&self, layout: core::alloc::Layout) -> Result<NonNull<u8>, ()> {
        if self.cur.is_null() {
            todo!()
        }
        let start = (*self.cur).pos.get();
        let offset = start.align_offset(layout.align());
        let start = start.add(offset);
        let end = start.add(layout.size());
        if end <= (*self.cur).end {
            (*self.cur).pos.set(end);
            Ok(NonNull::new_unchecked(start))
        } else {
            todo!()
        }
    }

    fn append_chunk(&self, min_layout: core::alloc::Layout) {

    }
}