use std::ptr::null;
use tonic::{Request, Response, Status};
use cosmossdk_core::c::{ModuleDescriptor, ModuleInitData};
use cosmossdk_core::KVStoreService;
use cosmossdk_core::tonic::{context};

include!(concat!(env!("OUT_DIR"), "/_includes.rs"));

struct Counter {
    kv_store_service: KVStoreService,
}

#[tonic::async_trait]
impl example::counter::v1::msg_server::Msg for Counter {
    async fn increment_counter(&self, request: Request<example::counter::v1::IncrementCounterRequest>) -> Result<Response<example::counter::v1::IncrementCounterResponse>, Status> {
        let mut store = self.kv_store_service.open(&mut context(&request));
        if let Some(val) = store.get(&[0]) {
            let mut val_be = u64::from_be_bytes(val[..8].try_into().unwrap());
            val_be += 1;
            store.set(&[0], val_be.to_be_bytes().as_ref());
            Ok(Response::new(example::counter::v1::IncrementCounterResponse {
                current: val_be,
            }))
        } else {
            let val_be = 1u64;
            store.set(&[0], val_be.to_be_bytes().as_ref());
            Ok(Response::new(example::counter::v1::IncrementCounterResponse {
                current: 1,
            }))
        }
    }
}

static MSG_SERVICE: example::counter::v1::msg_server::MsgServer<Counter> = example::counter::v1::msg_server::MsgServer::new(Counter {
    kv_store_service: todo!()
});

static PROTO_FILE_DESCRIPTORS: &'static [u8] = include_bytes!("file_descriptor_set.bin");

static MODULE_DESCRIPTORS: &'static [ModuleDescriptor] = &[
    ModuleDescriptor {
        name: "example.counter.v1".as_ptr(),
        name_len: "example.counter.v1".len(),
        // init_fn: fn
        init_fn: |init_data| { null() },
    },
];

static INIT_DATA: cosmossdk_core::c::InitData = cosmossdk_core::c::InitData {
    proto_file_descriptors: PROTO_FILE_DESCRIPTORS.as_ptr(),
    proto_file_descriptors_len: PROTO_FILE_DESCRIPTORS.len(),
    module_descriptors: MODULE_DESCRIPTORS.as_ptr(),
    num_modules: MODULE_DESCRIPTORS.len(),

};

#[no_mangle]
extern fn __init() -> *const cosmossdk_core::c::InitData {
    &INIT_DATA
}

