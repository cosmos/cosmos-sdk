use cosmossdk_core::Router;
use state_objects::UBigMap;
use zeropb::Server;
use crate::example::bank::v1::bank::{MsgSend, MsgSendResponse, MsgServer};

pub mod example {
    pub mod bank {
        pub mod v1 {
            pub mod bank {
                include!("example/bank/v1/bank.rs");
            }
        }
    }
}

pub struct Bank {
    balances: UBigMap<([u8], [u8])>,
}

impl MsgServer for Bank {
    fn send(&self, ctx: &mut ::zeropb::Context, req: &MsgSend) -> ::zeropb::Result<MsgSendResponse> {
        self.balances.safe_sub(ctx, todo!(), todo!()?;
        todo!()
    }
}

impl Router for Bank {
    fn route(&self, route_id: u64, ctx: usize, p0: usize, p1: usize) -> usize {
        <dyn MsgServer as Server>::route(self, route_id, todo!(), todo!(), todo!());
        todo!()
    }
}