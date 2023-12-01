use tonic::{Request, Response, Status};
use cosmossdk_core::KVStoreService;

pub mod example {
    pub mod counter {
        pub mod v1 {
            include!(concat!(env!("OUT_DIR"), "/example.counter.v1.rs"));
        }
    }
}

struct Counter {
    kv_store_service: KVStoreService
}

#[tonic::async_trait]
impl example::counter::v1::msg_server::Msg for Counter {
    async fn increment_counter(&self, request: Request<crate::example::counter::v1::IncrementCounterRequest>) -> Result<Response<crate::example::counter::v1::IncrementCounterResponse>, Status> {
        // let mut store = self.kv_store_service.open();
        todo!()
    }
}
