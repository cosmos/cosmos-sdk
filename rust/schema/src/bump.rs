use core::alloc::Layout;
use core::cell::Cell;
use core::cmp::max;
use core::ptr::NonNull;
use allocator_api2::alloc::{AllocError, Allocator};

// Very simple, custom bump allocator to avoid third party dependencies,
// reduce code size, and customize where chunks are allocated from and their sizes.
#[derive(Default)]
pub struct Bump {
    // the current chunk that is being allocated, if any
    cur: Cell<Option<NonNull<Footer>>>,
}

// a footer describing the chunk that is at the end of the chunk
struct Footer {
    // the start of the chunk that originally got allocated
    start: NonNull<u8>,
    // the current allocation position in the chunk
    pos: Cell<NonNull<u8>>,
    // a pointer to the footer of the previous chunk used in this allocator
    prev: Option<NonNull<Footer>>,
    // the layout of the chunk
    layout: Layout,
}

const FOOTER_SIZE: usize = core::mem::size_of::<Footer>();

unsafe impl Allocator for Bump {
    fn allocate(&self, layout: allocator_api2::alloc::Layout) -> Result<NonNull<[u8]>, AllocError> {
        unsafe {
            match self.cur.get() {
                None => {
                    const START_SIZE: usize = 4096;
                    self.alloc_chunk(START_SIZE, layout)
                }
                Some(mut footer) => {
                    // finding the starting allocation position
                    let pos = footer.as_ref().pos.get();
                    // align to layout.align()
                    let offset = pos.align_offset(layout.align());
                    // add offset to pos
                    let pos = pos.add(offset);
                    // compute the new position for allocation
                    let new_pos = pos.add(layout.size());
                    // check if the new position is before the footer in the chunk
                    if new_pos <= footer.cast() {
                        // update the position in the footer
                        footer.as_mut().pos.set(new_pos);
                        // return the allocated slice
                        Ok(NonNull::slice_from_raw_parts(pos, layout.size()))
                    } else {
                        self.alloc_chunk(footer.as_ref().layout.size(), layout)
                    }
                }
            }
        }
    }

    unsafe fn deallocate(&self, ptr: NonNull<u8>, layout: Layout) {
        // we don't need to deallocate, because this is a bump allocator
        // and we deallocate everything at once when the allocator is dropped
    }

    // TODO: attempt to extend the memory block in place
    // unsafe fn grow(&self, ptr: NonNull<u8>, old_layout: Layout, new_layout: Layout) -> Result<NonNull<[u8]>, AllocError> {
    //     todo!()
    // }
}

impl Bump {
    unsafe fn alloc_chunk(&self, start_size: usize, layout: Layout) -> Result<NonNull<[u8]>, AllocError> {
        let mut size = start_size;
        // the minimum size is the size of the layout plus the size of the footer
        let needed = layout.size() + FOOTER_SIZE;
        while size < needed {
            size <<= 1;
        }
        // we align to either he needed alignment or at least 16
        let align = max(layout.align(), 16);
        // the layout of the chunk
        let chunk_layout = Layout::from_size_align(size, align).map_err(|_| AllocError)?;
        // allocate the chunk, this will also be the newly allocated memory for the layout
        let start = NonNull::new_unchecked(alloc::alloc::alloc(chunk_layout));
        // update the allocation position
        let pos = start.add(layout.size());
        // find the end of the chunk
        let end = start.add(size);
        // the footer is at the end of the chunk
        let footer = end.sub(FOOTER_SIZE).cast::<Footer>();
        // TODO make sure the footer is at an aligned position
        assert_eq!(0, footer.align_offset(align_of::<Footer>()));
        // write the footer
        footer.write(Footer {
            start,
            pos: Cell::new(pos),
            // the previous footer is the current footer
            prev: self.cur.get(),
            layout: chunk_layout,
        });
        // update the current footer
        self.cur.set(Some(footer));
        Ok(NonNull::slice_from_raw_parts(start, layout.size()))
    }
}

impl Drop for Bump {
    fn drop(&mut self) {
        let mut maybe_footer = self.cur.get();
        while let Some(footer) = maybe_footer {
            let footer = unsafe { footer.as_ref() };
            maybe_footer = footer.prev;
            unsafe {
                alloc::alloc::dealloc(footer.start.as_ptr(), footer.layout);
            }
        }
    }
}

// struct Pool<const M: u32, const STEPS: usize> {
//     head_by_size: [Cell<Option<NonNull<Footer>>>; STEPS],
// }
//
// impl<const M: u32, const N: usize> Pool<M, N> {
//     const START_SIZE: usize = 2usize.pow(M);
//     unsafe fn alloc_chunk(&self, step: usize) -> Result<NonNull<Footer>, AllocError> {
//         if step >= N {
//             return Err(AllocError);
//         }
//         if let Some(mut footer) = self.head_by_size[step].get() {
//             let f = footer.as_mut();
//             self.head_by_size[step].set(footer.as_ref().prev.get());
//             f.pos.set(f.start);
//             f.prev.set(None);
//             Ok(footer)
//         } else {
//             let size = Self::START_SIZE << step;
//             let layout = Layout::from_size_align(size, size).map_err(|_| AllocError)?;
//             let ptr = unsafe { NonNull::new_unchecked(alloc::alloc::alloc(layout)) };
//             let end = ptr.add(size);
//             let footer = ptr.sub(FOOTER_SIZE).cast::<Footer>();
//             footer.write(Footer {
//                 start: ptr,
//                 pos: Cell::new(ptr),
//                 prev: Cell::new(None),
//                 size: size,
//             });
//             Ok(footer)
//         }
//     }
//
//     unsafe fn dealloc_chunk(&self, footer: NonNull<Footer>) -> Result<(), ()> {
//         // find step size
//         // put at head of list of that size
//         todo!()
//     }
// }