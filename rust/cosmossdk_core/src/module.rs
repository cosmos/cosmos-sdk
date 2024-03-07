use crate::router::Router;

pub trait Module: Router + ModuleDescriptor {
}

pub trait ModuleDescriptor {
}

pub trait ModuleBundle: Router + ModuleBundleDescriptor {
    fn route(&self, route_id: u64, caller: u64, ctx: &mut zeropb::Context, req: *const u8) -> Result<*const u8, u32>;
}

pub trait ModuleBundleDescriptor {
}


