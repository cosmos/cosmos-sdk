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
    register_unary_method: unsafe extern "C" fn(service: *const u8, name: *const u8, handler: MethodIn1Out1),
    callbacks: *const CallbackAPI,
}

pub(crate) type InvokeIn1Out1 = unsafe extern "C" fn(id: usize, ctx: usize, arg1: *const u8, arg1_len: usize, res: **mut u8, res_len: *mut usize) -> i64;

#[repr(C)]
struct CallbackAPI {
    invoke_unary: InvokeUnary,
    method0: Method0,
    method_in1: MethodIn1,
    method_in2: MethodIn2,
    method_in1_out1: MethodIn1Out1,
    method_in2_out2: MethodIn2Out2,
}

// next, close
pub(crate) type Method0 = unsafe extern "C" fn(ctx: usize) -> i64;
// has, delete
pub(crate) type MethodIn1 = unsafe extern "C" fn(ctx: usize, arg1: *const u8, arg1_len: usize) -> i64;
// set
pub(crate) type MethodIn2 = unsafe extern "C" fn(ctx: usize, arg1: *const u8, arg1_len: usize, arg2: *const u8, arg2_len: usize) -> i64;
// get, unary method
pub(crate) type MethodIn1Out1 = unsafe extern "C" fn(ctx: usize, arg1: *const u8, arg1_len: usize, res: **mut u8, res_len: *mut usize) -> i64;
// iterate, iterate_reverse
pub(crate) type MethodIn2Out2 = unsafe extern "C" fn(ctx: usize, arg1: *const u8, arg1_len: usize, arg2: *const u8, arg2_len: usize, res: **mut u8, res_len: *mut usize, res2: **mut u8, res2_len: **mut u8) -> i64;
