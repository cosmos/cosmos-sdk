use core::ptr::NonNull;
use bump_scope::{Bump, BumpBox, BumpString, BumpVec};

pub struct MemoryManager<'b, 'a: 'b> {
    handles: BumpVec<'b, 'a, NonNull<dyn DeferDrop + 'b>>,
}

impl<'b, 'a: 'b> MemoryManager<'b, 'a> {
    pub fn new_scope(bump: &Bump) -> MemoryManager<'a, 'a> {
        MemoryManager {
            handles: BumpVec::new_in(bump),
        }
    }

    pub fn scope(&self) -> &'b bump_scope::BumpScope<'a> {
        self.handles.bump()
    }

    pub fn unpack_slice<T>(&mut self, vec: BumpVec<'b, 'a, T>) -> &'b [T] {
        unsafe {
            let b = vec.into_boxed_slice();
            let slice = b.as_non_null_slice().as_ptr() as *const [T];
            struct Dropper<'a, U> {
                b: BumpBox<'a, [U]>,
            }
            let dropper = self.scope().alloc(Dropper { b });
            self.handles.push(dropper.into_raw() as NonNull<dyn DeferDrop + 'b>);
            &*slice
        }
    }
}

impl<'b, 'a: 'b> Drop for MemoryManager<'b, 'a> {
    fn drop(&mut self) {
        for handle in self.handles.drain(..) {
            unsafe {
                handle.as_ptr().drop_in_place();
            }
        }
    }
}

trait DeferDrop {}
impl<T> DeferDrop for T {}