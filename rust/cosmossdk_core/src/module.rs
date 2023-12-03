// use crate::c::{Method0, MethodIn1, MethodIn1Out1, MethodIn2, MethodIn2Out2, MethodUnary};
use crate::services::ServiceBundle;

pub trait Module {
    type Config;

    fn init(config: Self::Config, resolver: &Resolver) -> Self;

    fn register(&self, registrar: &mut Registrar) -> Result<(), zeropb::Code>;
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


type EncodingType = u32;
const ENCODING_CUSTOM: EncodingType = 0;
const ENCODING_ZEROPB: EncodingType = 1;
const ENCODING_PROTO_BINARY: EncodingType = 2;

