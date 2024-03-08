// mod async;

use dashu_int::UBig;
use cosmossdk_core::{BeginWriteContext, Context, ReadContext, WriteContext};

pub trait KeyCodec {
    type T;
}

pub trait ValueCodec {
    type T;
}

impl KeyCodec for u64 {
    type T = u64;
}

impl KeyCodec for [u8] {
    type T = [u8];
}

impl <P1:KeyCodec, P2:KeyCodec> KeyCodec for (P1, P2) {
    type T = (P1::T, P2::T);
}

impl <P1:KeyCodec, P2:KeyCodec, P3: KeyCodec> KeyCodec for (P1, P2, P3) {
    type T = (P1::T, P2::T, P3::T);
}

struct CompactU64;

impl KeyCodec for CompactU64 {
    type T = u64;
}

struct Map<K, V> {}

impl <K:KeyCodec, V> Map<K, V> {
    fn get(&self, key: K::T) -> Option<V::T> { todo!() }
}

struct MyModule {
    myMap: Map<CompactU64, u64>
}

impl ValueCodec for UBig {
    type T = UBig;
}

pub struct UBigMap<K> {}

impl <K:KeyCodec> UBigMap<K> {
    pub fn has(&self, ctx: &Context, key: K::T) ->cosmossdk_core::Result<bool> {
        todo!()
    }

    pub fn read(&self, ctx: &Context, key: K::T) ->cosmossdk_core::Result<UBig> {
        todo!()
    }

    pub fn delete(&self, ctx: &mut Context, key: K::T) -> cosmossdk_core::Result<()> {
        todo!()
    }

    pub fn safe_sub(&self, ctx: &mut Context, key: K::T, value: UBig) -> cosmossdk_core::Result<UBig> {
        todo!()
    }

    pub fn add(&self, ctx: &mut Context, key: K::T, value: UBig) {
        todo!()
    }
}

