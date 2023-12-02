#[repr(C)]
struct InitData {
    proto_file_descriptors: *const u8,
    proto_file_descriptors_len: u32,
    module_names: *const *const u8,
    module_names_len: u32,
    module_init_fns: *const ModuleInitFn,
}

type ModuleInitFn = unsafe extern "C" fn(init_data: *const ModuleInitData) -> i32;

#[repr(C)]
struct ModuleInitData {
    config: *const u8,
    config_len: u32,
    register_unary_method: unsafe extern "C" fn(name: *const u8, handler: UnaryMethodHandler),
}

type UnaryMethodHandler = unsafe extern "C" fn(ctx: u32, req: *const u8, res: *mut u8) -> i32;

#[repr(C)]
struct StoreAPI {
    open: OpenFn,
    has: HasFn,
    get: GetFn,
    set: SetFn,
    delete: DeleteFn,
}

type OpenFn = unsafe extern "C" fn(module: usize, context: usize) -> usize;
type HasFn = unsafe extern "C" fn(store: usize, key: *const u8, len: u32) -> u32;
type GetFn = unsafe extern "C" fn(store: usize, key: *const u8, len: u32, value: *mut u8, vlen: *mut u32) -> u32;
type SetFn = unsafe extern "C" fn(store: usize, key: *const u8, len: u32, value: *const u8, vlen: u32) -> u32;
type DeleteFn = unsafe extern "C" fn(store: usize, key: *const u8, len: u32) -> u32;

