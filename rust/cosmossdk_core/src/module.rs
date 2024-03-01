pub trait Module {
    // TODO: descriptor
    // TODO: init
    fn route_tx(&self, method_id: u32, ctx: &mut zeropb::Context, req: *const u8) -> Result<*const u8, u32>;
    fn route_query(&self, method_id: u32, ctx: &zeropb::Context, req: *const u8) -> Result<*const u8, u32>;
    fn route_internal(&self, method_id: u32, caller: u64, ctx: &mut zeropb::Context, req: *const u8) -> Result<*const u8, u32>;
}

pub trait ModuleBundle {
    fn route(&self, route_id: u64, caller: u64, ctx: &mut zeropb::Context, req: *const u8) -> Result<*const u8, u32>;
}


