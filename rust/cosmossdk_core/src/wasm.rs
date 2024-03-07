#![no_std]

/// cbindgen:ignore
#[link(wasm_import_module = "CosmosSDK")]
extern "C" {
    fn invoke(ctx: u32, method_id: u32, req: *const u8, res: *mut u8) -> i32;
}

