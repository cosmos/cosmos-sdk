// mod async;

use dashu_int::UBig;
use cosmossdk_core::{Context};

pub trait KeyCodec {
    type In<'a>;
}

pub trait ValueCodec {
    type In;
    type Out;
}

// impl KeyCodec for u64 {
//     type In = u64;
// }

pub struct Bytes;

impl KeyCodec for Bytes {
    type In<'a> = &'a [u8];
}

pub struct Pair<P1, P2>(pub P1, pub P2);

impl <P1:KeyCodec, P2:KeyCodec> KeyCodec for Pair<P1, P2> {
    type In<'a> = Pair<P1::In<'a>, P2::In<'a>>;
}

pub struct Str;

impl KeyCodec for Str {
    type In<'a> = &'a str;
}

impl ValueCodec for bool {
    type In = bool;
    type Out = bool;
}

// impl <P1:KeyCodec, P2:KeyCodec> KeyCodec for (P1, P2) {
//     type In<'a> = (P1::In<'a>, P2::In<'a>);
// }
//
// impl <P1:KeyCodec, P2:KeyCodec, P3: KeyCodec> KeyCodec for (P1, P2, P3) {
//     type In = (P1::In, P2::In, P3::In);
// }

// struct CompactU64;
//
// impl KeyCodec for CompactU64 {
//     type In = u64;
// }

pub struct Map<K, V> {}

impl <K:KeyCodec, V: ValueCodec> Map<K, V> {
    pub fn get(&self, ctx: &cosmossdk_core::Context, key: &K::In) -> cosmossdk_core::Result<&V::Out> {
        todo!()
    }

    pub fn get_last_block(&self, ctx: &cosmossdk_core::Context, key: &K::In) -> cosmossdk_core::Result<&V::Out> {
        todo!()
    }

    pub fn set(&self, ctx: &cosmossdk_core::Context, key: K::In, value: &V::In) -> cosmossdk_core::Result<()> {
        todo!()
    }
}
//
// struct MyModule {
//     myMap: Map<CompactU64, u64>
// }
//
impl ValueCodec for UBig {
    type In = UBig;
    type Out = UBig;
}

pub struct UBigMap<K> {
    _k: std::marker::PhantomData<K>
}

impl <K:KeyCodec> UBigMap<K> {
    pub fn has(&self, ctx: &Context, key: &K::In<'_>) ->cosmossdk_core::Result<bool> {
        todo!()
    }

    pub fn read(&self, ctx: &Context, key: &K::In<'_>) ->cosmossdk_core::Result<UBig> {
        todo!()
    }

    pub fn delete(&self, ctx: &mut Context, key: &K::In<'_>) -> cosmossdk_core::Result<()> {
        todo!()
    }

    pub fn safe_sub(&self, ctx: &mut Context, key: &K::In<'_>, value: &UBig) -> cosmossdk_core::Result<UBig> {
        todo!()
    }

    pub fn add(&self, ctx: &mut Context, key: &K::In<'_>, value: &UBig) {
        todo!()
    }
}
