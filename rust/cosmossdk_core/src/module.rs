// use crate::c::{Method0, MethodIn1, MethodIn1Out1, MethodIn2, MethodIn2Out2, MethodUnary};
use crate::services::ServiceBundle;

pub trait Module {
    // TODO: descriptor
    // TODO: init
    fn route(&self, method_id: u32, ctx: &mut zeropb::Context, req: *const u8) -> Result<*const u8, u32>;
}

pub struct Registrar {}

impl Registrar {
    // pub fn register_unary(&mut self, service_name: &str, method_name: &str, encoding: EncodingType, handler: MethodIn1Out1) {
    // }
}

pub struct Resolver {}

impl Resolver {
    // pub fn resolve_unary(&self, service_name: &str, method_name: &str) -> MethodUnary {
    //     todo!()
    // }
    //
    // pub fn resolve0(&self, service_name: &str, method_name: &str) -> Method0 {
    //     todo!()
    // }
    //
    // pub fn resolve_in1(&self, service_name: &str, method_name: &str) -> MethodIn1 {
    //     todo!()
    // }
    //
    // pub fn resolve_in1_out1(&self, service_name: &str, method_name: &str) -> MethodIn1Out1 {
    //     todo!()
    // }
    //
    // pub fn resolve_in2(&self, service_name: &str, method_name: &str) -> MethodIn2 {
    //     todo!()
    // }
    //
    // pub fn resolve_in2_out2(&self, service_name: &str, method_name: &str) -> MethodIn2Out2 {
    //     todo!()
    // }
}



