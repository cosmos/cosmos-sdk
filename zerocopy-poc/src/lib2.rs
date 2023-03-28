use alloc::slice::from_raw_parts;

#[repr(C)]
struct RootInner {
    buf: *mut u8,
    capacity: usize,
    large_buffers: *mut *mut u8,
    large_buffers_capacity: usize,
}

struct Root<T> {
    inner: RootInner,
    size: usize,
}

trait RootLike {}

impl<T> Root<T> {
    fn alloc(&mut self, size: usize) -> (*mut u8, u16) {
        panic!("TODO")
    }

    fn load(&mut self, offset: u16) -> *mut u8 {
        panic!("TODO")
    }

    fn new_string(&mut self, s: &str) -> String {
        let size = s.len();
        let (ptr, offset) = self.alloc(size);
        unsafe {
            std::ptr::copy_nonoverlapping(s.as_ptr(), ptr, size);
        }
        String { offset }
    }
}

#[repr(C, align(1))]
struct String {
    offset: u16,
}

impl String {
    fn get(&self, root: &dyn RootLike) -> &str {
        let offset = root.load(self.offset);
        // alloc::slice::from_raw_parts
        panic!("TODO")
    }
}

type Context = *const u8;

trait BeginBlocker {
    fn begin_block(&self, ctx: &Context);
}

trait EndBlocker {
    fn end_block(&self, ctx: &Context);
}

trait GenesisHandler {
    fn default_genesis(&self, ctx: &Context);
    fn init_genesis(&self, ctx: &Context);
    fn export_genesis(&self, ctx: &Context);
    fn validate_genesis(&self, ctx: &Context);
}
