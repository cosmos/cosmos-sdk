extern crate core;
use core::todo;
use crate::{Client, ClientConnection, Code, Context, Result, Router, Server};
use crate::raw::{RawBox, RawBytes};
use core::result::{Result::{Err, Ok}};

#[cfg(feature="alloc")]
use crate::sync::{Completer, Completer1, PrepareContext};

pub struct StoreClient {
    conn: RawBox<dyn ClientConnection>,
    route_id: u64,
}

#[cfg_attr(any(test, feature = "test-util"), mockall::automock)]
pub trait Store {
    fn get(&self, ctx: &mut Context, key: &[u8]) -> Result<RawBytes>;

    fn set(&self, ctx: &mut Context, key: &[u8], value: &[u8]) -> Result<()>;

    fn delete(&self, ctx: &mut Context, key: &[u8]) -> Result<()>;

    fn has(&self, ctx: &mut Context, key: &[u8]) -> Result<bool>;

    fn get_stale(&self, ctx: &mut Context, key: &[u8]) -> Result<RawBytes>;

    fn set_lazy(&self, ctx: &mut Context, key: &[u8], value_fn: fn(&[u8]) -> RawBytes) -> Result<()>;

    #[cfg(feature="alloc")]
    fn prepare_get(&self, ctx: &PrepareContext, key: &[u8]) -> Result<Completer<RawBytes>> {
        Ok(alloc::boxed::Box::new(move |ctx| self.get(ctx, key)))
    }

    #[cfg(feature="alloc")]
    fn prepare_set(&self, ctx: &PrepareContext, key: &[u8]) -> Result<Completer1<RawBytes, ()>> {
        Ok(alloc::boxed::Box::new(move |ctx, value| self.set(ctx, key, value)))
    }

    #[cfg(feature="alloc")]
    fn prepare_delete(&self, ctx: &PrepareContext, key: &[u8]) -> Result<Completer<()>> {
        Ok(alloc::boxed::Box::new(move |ctx| self.delete(ctx, key)))
    }

    #[cfg(feature="alloc")]
    fn prepare_has(&self, ctx: &PrepareContext, key: &[u8]) -> Result<Completer<bool>> {
        Ok(alloc::boxed::Box::new(move |ctx| self.has(ctx, key)))
    }

    #[cfg(feature="alloc")]
    fn prepare_get_stale(&self, ctx: &PrepareContext, key: &[u8]) -> Result<Completer<RawBytes>> {
        Ok(alloc::boxed::Box::new(move |ctx| self.get_stale(ctx, key)))
    }

    #[cfg(feature="alloc")]
    fn prepare_set_lazy(&self, ctx: &PrepareContext, key: &[u8]) -> Result<Completer1<fn(&[u8]) -> RawBytes, ()>> {
        Ok(alloc::boxed::Box::new(move |ctx, value_fn| self.set_lazy(ctx, key, value_fn)))

    }
}

impl Router for dyn Store {

}

impl Server for dyn Store {

}

impl Client for StoreClient {
    fn new(route_id: u64, conn: RawBox<dyn ClientConnection>) -> Self {
        StoreClient {
            conn,
            route_id,
        }
    }
}

impl StoreClient {
    fn get(&self, ctx: &mut Context, key: &[u8]) -> Result<RawBytes> {
        self.conn.route_io(self.route_id & 0x1, ctx, key)
    }

    fn set(&self, ctx: &mut Context, key: &[u8], value: &[u8]) -> Result<()> {
        self.conn.route_i2(self.route_id & 0x2, ctx, key, value)
    }

    fn delete(&self, ctx: &mut Context, key: &[u8]) -> Result<()> {
        self.conn.route_i1(self.route_id & 0x3, ctx, key)
    }

    fn has(&self, ctx: &mut Context, key: &[u8]) -> Result<bool> {
        match self.conn.route_io(self.route_id & 0x4, ctx, key) {
            Ok(_) => Ok(true),
            Err(e) => {
                if e.code == Code::NotFound {
                    Ok(false)
                } else {
                    Err(e)
                }
            }
        }

    }

    fn get_stale(&self, ctx: &mut Context, key: &[u8]) -> Result<RawBytes> {
        self.conn.route_io(self.route_id & 0x5, ctx, key)
    }

    fn set_lazy(&self, ctx: &mut Context, key: &[u8], value_fn: fn(&[u8]) -> RawBytes) -> Result<()> {
        // self.conn.route_i2(self.route_id & 0x6, ctx, key, value_fn(key))
        todo!()
    }
}
