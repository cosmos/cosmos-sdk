use crate::context::{Context};
use crate::ffi;

pub struct KVStoreService {
    module_id: u32,
}

#[cfg(not(target_arch = "wasm32"))]
impl KVStoreService {
    pub fn open(&self, ctx: &Context) -> KVStore {
        let store_id = unsafe { ffi::cosmossdk_store_open(self.module_id, ctx.id) };
        KVStore { store_id }
    }

    pub fn open_mut(&self, ctx: &mut Context) -> KVStore {
        let store_id = unsafe { ffi::cosmossdk_store_open(self.module_id, ctx.id) };
        KVStore { store_id }
    }
}

pub struct KVStore {
    store_id: u32,
}

#[cfg(not(target_arch = "wasm32"))]
impl KVStore {
    pub fn has(&self, key: &[u8]) -> bool {
        unsafe { ffi::cosmossdk_store_has(self.store_id, key.as_ptr() as *mut u8, key.len()) == 0 }
    }

    pub fn get(&self, key: &[u8]) -> Option<Vec<u8>> {
        let value = vec![0u8; 0x10000];
        let mut value_len = value.len();
        let res = unsafe {
            ffi::cosmossdk_store_get(
                self.store_id,
                key.as_ptr(),
                key.len(),
                value.as_ptr() as *mut u8,
                &mut value_len,
            )
        };
        if res == 0 {
            Some(value[..value_len].to_vec())
        } else {
            None
        }
    }

    pub fn set(&mut self, key: &[u8], value: &[u8]) {
        unsafe {
            ffi::cosmossdk_store_set(
                self.store_id,
                key.as_ptr(),
                key.len(),
                value.as_ptr(),
                value.len(),
            )
        };
    }

    pub fn delete(&mut self, key: &[u8]) {
        unsafe { ffi::cosmossdk_store_delete(self.store_id, key.as_ptr() as *mut u8, key.len()) };
    }
}

impl Drop for KVStore {
    fn drop(&mut self) {
        unsafe { ffi::cosmossdk_store_close(self.store_id) };
    }
}