extern crate core;
use core::result::{Result::Err};
use crate::{Code, Context, Result};
use crate::Code::Unimplemented;
use crate::raw::{RawBox, RawBytes};
use core::convert::{From, Into};

pub trait Router {
    // fn route(&self, method_id: u64, ctx: &mut Context, req: *mut u8, res: *mut *mut u8) -> Code { Unimplemented }
    fn route_io(&self, method_id: u64, ctx: &mut Context, req: &[u8]) -> Result<RawBytes> { Err(Unimplemented.into()) }
    fn route_i1(&self, method_id: u64, ctx: &mut Context, p1: &[u8]) -> Result<()> { Err(Unimplemented.into()) }
    fn route_i2(&self, method_id: u64, ctx: &mut Context, p1: &[u8], p2: &[u8]) -> Result<()> { Err(Unimplemented.into()) }
    // fn route_method(&self, method: &str, ctx: &mut Context, req: &[u8]) -> Result<RawBytes> { Err(Unimplemented.into()) }
    // fn route_service_method(&self, svc: &str, method: &str, ctx: &mut Context, req: &[u8]) -> Result<RawBytes> { Err(Unimplemented.into()) }
}

pub trait Server: Router {
    // fn describe(descriptor: &mut crate::types::cosmos::core::v1alpha1::bundle::ModuleOutput) -> zeropb::Result<()>;
    // fn route_io(&self, method_id: u64, ctx: &mut Context, req: &[u8]) -> Result<RawBytes> { Err(Unimplemented.into()) }
    // fn route_i1(&self, method_id: u64, ctx: &mut Context, p1: &[u8]) -> Result<()> { Err(Unimplemented.into()) }
    // fn route_i2(&self, method_id: u64, ctx: &mut Context, p1: &[u8], p2: &[u8]) -> Result<()> { Err(Unimplemented.into()) }
}

pub trait ClientConnection: Router {
    // fn route_id(&self) -> u64 { 0 }
    // fn route_service_method(&self, method_id: u64, service_name: &str, method: &str, ctx: &mut Context, req: &[u8]) -> Result<RawBytes>;
}

pub trait Client {
    fn new(route_id: u64, conn: RawBox<dyn ClientConnection>) -> Self;
}
