#[repr(C)]
pub struct InitData {
    pub proto_file_descriptors: *const u8,
    pub proto_file_descriptors_len: usize,
    pub module_descriptors: *const ModuleDescriptor,
    pub num_modules: usize,
}

#[repr(C)]
pub struct ModuleDescriptor {
    pub name: *const u8,
    pub name_len: usize,
    // pub init_fn: ModuleInitFn,
}

unsafe impl Sync for ModuleDescriptor {}
unsafe impl Send for ModuleDescriptor {}

unsafe impl Sync for InitData {}
unsafe impl Send for InitData {}

#[repr(C)]
pub struct ModuleInitData {
    pub config: *const u8,
    pub config_len: u32,
    pub register_unary_method: extern "C" fn(service: *const u8, service_len: usize, method: *const u8, method_len: usize, encoding: EncodingType, handler: UnaryMethodHandler) -> u32,
}

unsafe impl Sync for ModuleInitData {}
unsafe impl Send for ModuleInitData {}

pub type ModuleInitFn = extern "C" fn(init_data: *const ModuleInitData) -> *const ();

pub type UnaryMethodHandler = unsafe extern "C" fn(ctx: u32, req: *const u8, req_len: usize, res: *mut u8, res_len: *mut usize) -> u32;

pub type InitFn = extern "C" fn() -> *const InitData;

type EncodingType = u32;
const ENCODING_CUSTOM: EncodingType = 0;
const ENCODING_ZEROPB: EncodingType = 1;
const ENCODING_PROTO_BINARY: EncodingType = 2;

#[cfg(feature = "example")]
#[no_mangle]
pub extern fn __init() -> *const InitData {
    null()
}

// #[repr(C)]
// struct InitData {
//     proto_file_descriptors: *const u8,
//     proto_file_descriptors_len: u32,
//     module_names: *const *const u8,
//     module_names_len: u32,
//     module_init_fns: *const ModuleInitFn,
// }
//
// type ModuleInitFn = unsafe extern "C" fn(init_data: *const ModuleInitData) -> i32;
//
// #[repr(C)]
// struct ModuleInitData {
//     config: *const u8,
//     config_len: u32,
//     register_unary_method: unsafe extern "C" fn(service: *const u8, name: *const u8, handler: MethodIn1Out1),
//     callbacks: *const CallbackAPI,
// }

// pub(crate) type InvokeIn1Out1 = unsafe extern "C" fn(id: usize, ctx: usize, arg1: *const u8, arg1_len: usize, res: **mut u8, res_len: *mut usize) -> i64;

// #[repr(C)]
// struct CallbackAPI {
//     invoke_unary: InvokeUnary,
//     method0: Method0,
//     method_in1: MethodIn1,
//     method_in2: MethodIn2,
//     method_in1_out1: MethodIn1Out1,
//     method_in2_out2: MethodIn2Out2,
// }

// // next, close
// pub(crate) type Method0 = unsafe extern "C" fn(ctx: usize) -> i64;
// // has, delete
// pub(crate) type MethodIn1 = unsafe extern "C" fn(ctx: usize, arg1: *const u8, arg1_len: usize) -> i64;
// // set
// pub(crate) type MethodIn2 = unsafe extern "C" fn(ctx: usize, arg1: *const u8, arg1_len: usize, arg2: *const u8, arg2_len: usize) -> i64;
// // get, unary method
// pub(crate) type MethodIn1Out1 = unsafe extern "C" fn(ctx: usize, arg1: *const u8, arg1_len: usize, res: **mut u8, res_len: *mut usize) -> i64;
// // iterate, iterate_reverse
// pub(crate) type MethodIn2Out2 = unsafe extern "C" fn(ctx: usize, arg1: *const u8, arg1_len: usize, arg2: *const u8, arg2_len: usize, res: **mut u8, res_len: *mut usize, res2: **mut u8, res2_len: **mut u8) -> i64;
