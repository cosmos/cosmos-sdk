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
