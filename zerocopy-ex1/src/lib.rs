#[repr(C)]
pub struct Foo {
    pub a: u32,
    pub b: u16,
    pub c: u64,
}

#[no_mangle]
pub extern fn baz(left: usize, right: *mut u8) -> usize {
    left
}

#[no_mangle]
pub extern fn foo(bar: Bar) -> Foo {
    Foo { a: 1, b: 2, c: 3 }
}

#[repr(u8)]
pub enum Bar {
    A,
    B,
    C,
}

#[no_mangle]
pub extern fn cosmos_register_modules(register: RegisterFn) -> i32 {
    0
}

#[no_mangle]
pub extern fn call(cb: fn(i32) -> i32) -> i32 {
    cb(1)
}

type RegisterFn = extern fn(module_info_size: u32, module_info_data: *const u8, providers: *const ProviderFn) -> i32;

type ProviderFn = extern fn(config_size: u32, config_data: *const u8, inputs: *const usize, register_output: RegisterOutputFn, err: *mut u8) -> i32;

type RegisterOutputFn = extern fn(output: *const usize);

// wasm stuff:

#[no_mangle]
pub extern fn cosmos_wasm_register_modules() -> i32 {
    unsafe { cosmos_wasm_register(0, std::ptr::null(), std::ptr::null()) }
}

#[no_mangle]
pub extern fn cosmos_wasm_init(provider_id: i32, config_size: u32, config_data: *const u8, err: *mut u8) -> i32 {
    0
}

extern "C" {
    fn cosmos_wasm_register(module_info_size: u32, module_info_data: *const u8, providers: *const i32) -> i32;
}

