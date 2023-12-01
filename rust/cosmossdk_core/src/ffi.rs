struct ModuleInitializers {
    count: usize,
    initializers: *const ModuleInitializer,
}

struct ModuleInitializer {
    name: *const u8,
    proto_files: *const *const u8,
}
#[repr(C)]
pub struct GoSlice {
    data: *mut u8,
    len: isize,
    cap: isize,
}

extern "C" {
    pub fn cosmossdk_store_open(module_id: u32, ctx_id: u32) -> u32;

    pub fn cosmossdk_store_has(store_id: u32, key_ptr: *mut u8, key_len: usize) -> u32;

    pub fn cosmossdk_store_get(
        store_id: u32,
        key_ptr: *const u8,
        key_len: usize,
        value_ptr: *mut u8,
        value_len: *mut usize
    ) -> i32;

    pub fn cosmossdk_store_set(
        store_id: u32,
        key_ptr: *const u8,
        key_len: usize,
        value_ptr: *const u8,
        value_len: usize
    ) -> i32;

    pub fn cosmossdk_store_delete(store_id: u32, key_ptr: *mut u8, key_len: usize) -> i32;

    pub fn cosmossdk_store_close(store_id: u32) -> i32;
}