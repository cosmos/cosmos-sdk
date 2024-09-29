use alloc::string::String;
use bump_scope::{BumpScope, BumpString};
use crate::decoder::{decode, DecodeError};
use crate::list::ListVisitor;
use crate::mem::MemoryManager;
use crate::structs::StructDecodeVisitor;
use crate::value::Value;

pub fn decode_value<'b, 'a: 'b, V: Value<'a>>(input: &'a [u8], memory_manager: &'b MemoryManager<'a>) -> Result<V, DecodeError> {
    let mut decoder = Decoder { buf: input, scope: memory_manager };
    decode(&mut decoder)
}

struct Decoder<'b, 'a: 'b> {
    buf: &'a [u8],
    scope: &'b MemoryManager<'a>,
}

impl<'b, 'a: 'b> Decoder<'b, 'a> {
    fn read_bytes(&mut self, size: usize) -> Result<&'a [u8], DecodeError> {
        if self.buf.len() < size {
            return Err(DecodeError::OutOfData);
        }
        let bz = &self.buf[0..size];
        self.buf = &self.buf[size..];
        Ok(bz)
    }
}

impl<'b, 'a: 'b> crate::decoder::Decoder<'a> for Decoder<'b, 'a> {
    fn decode_u32(&mut self) -> Result<u32, DecodeError> {
        let bz = self.read_bytes(4)?;
        Ok(u32::from_le_bytes(bz.try_into().unwrap()))
    }

    fn decode_u128(&mut self) -> Result<u128, DecodeError> {
        let bz = self.read_bytes(16)?;
        Ok(u128::from_le_bytes(bz.try_into().unwrap()))
    }

    fn decode_borrowed_str(&mut self) -> Result<&'a str, DecodeError> {
        let bz = self.buf;
        self.buf = &[];
        Ok(core::str::from_utf8(bz).map_err(|_| DecodeError::InvalidData)?)
    }

    fn decode_owned_str(&mut self) -> Result<String, DecodeError> {
        let bz = self.buf;
        self.buf = &[];
        Ok(String::from_utf8(bz.to_vec()).map_err(|_| DecodeError::InvalidData)?)
    }

    fn decode_struct<V: StructDecodeVisitor<'a>>(&mut self, visitor: &mut V) -> Result<(), DecodeError> {
        let mut i = 0;
        let mut sub = Decoder { buf: self.buf, scope: self.scope };
        let mut inner = InnerDecoder { outer: &mut sub };
        for _ in V::FIELDS.iter() {
            visitor.decode_field(i, &mut inner)?;
            i += 1;
        }
        Ok(())
    }

    fn decode_list<T, V: ListVisitor<'a, T>>(&mut self, visitor: &mut V) -> Result<(), DecodeError> {
        let size = self.decode_u32()? as usize;
        visitor.init(size, &self.scope)?;
        let mut sub = Decoder { buf: self.buf, scope: self.scope };
        let mut inner = InnerDecoder { outer: &mut sub };
        for _ in 0..size {
            visitor.next(&mut inner)?;
        }
        Ok(())
    }

    fn mem_manager(&self) -> &MemoryManager<'a> {
        &self.scope
    }
}

struct InnerDecoder<'b, 'a: 'b> {
    outer: &'b mut Decoder<'b, 'a>,
}
impl<'b, 'a: 'b> crate::decoder::Decoder<'a> for InnerDecoder<'b, 'a> {
    fn decode_u32(&mut self) -> Result<u32, DecodeError> {
        self.outer.decode_u32()
    }

    fn decode_u128(&mut self) -> Result<u128, DecodeError> {
        self.outer.decode_u128()
    }

    fn decode_borrowed_str(&mut self) -> Result<&'a str, DecodeError> {
        let size = self.decode_u32()? as usize;
        let bz = self.outer.read_bytes(size)?;
        Ok(core::str::from_utf8(bz).map_err(|_| DecodeError::InvalidData)?)
    }

    fn decode_owned_str(&mut self) -> Result<String, DecodeError> {
        let size = self.decode_u32()? as usize;
        let bz = self.outer.read_bytes(size)?;
        Ok(String::from_utf8(bz.to_vec()).map_err(|_| DecodeError::InvalidData)?)
    }

    fn decode_struct<V: StructDecodeVisitor<'a>>(&mut self, visitor: &mut V) -> Result<(), DecodeError> {
        let size = self.decode_u32()? as usize;
        let bz = self.outer.read_bytes(size)?;
        let mut sub = Decoder { buf: bz, scope: self.outer.scope };
        sub.decode_struct(visitor)
    }

    fn decode_list<T, V: ListVisitor<'a, T>>(&mut self, visitor: &mut V) -> Result<(), DecodeError> {
        todo!()
    }

    fn mem_manager(&self) -> &MemoryManager<'a> {
        &self.outer.scope
    }
}

#[cfg(test)]
mod tests {
    extern crate std;

    use alloc::vec;
    use bump_scope::{Bump, BumpScope};
    use crate::binary::decoder::decode_value;
    use crate::binary::encoder::encode_value;
    use crate::decoder::{DecodeError, Decoder};
    use crate::encoder::{EncodeError, Encoder};
    use crate::field::Field;
    use crate::mem::MemoryManager;
    use crate::structs::{StructDecodeVisitor, StructEncodeVisitor, StructSchema};
    use crate::types::{to_field, StrT, StructT, UIntNT};
    use crate::value::Value;

    #[test]
    fn test_u32_decode() {
        let buf: [u8; 4] = [10, 0, 0, 0];
        let bump = Bump::new();
        let mut mem = MemoryManager::new(bump.as_scope());
        let x = decode_value::<u32>(&buf, &mut mem).unwrap();
        assert_eq!(x, 10);
    }

    #[test]
    fn test_decode_borrowed_string() {
        let str = "hello";
        let bump = Bump::new();
        let mut mem = MemoryManager::new(bump.as_scope());
        let x = decode_value::<&str>(str.as_bytes(), &mut mem).unwrap();
        assert_eq!(x, "hello");
    }

    #[test]
    fn test_decode_owned_string() {
        let str = "hello";
        let bump = Bump::new();
        let mut mem = MemoryManager::new(bump.as_scope());
        let x = decode_value::<alloc::string::String>(str.as_bytes(), &mut mem).unwrap();
        assert_eq!(x, "hello");
    }

    #[derive(Debug, PartialEq)]
    struct Coin<'a> {
        denom: &'a str,
        amount: u128,
    }

    unsafe impl<'a> StructSchema for Coin<'a> {
        const FIELDS: &'static [Field<'static>] = &[
            to_field::<StrT>().with_name("denom"),
            to_field::<UIntNT<16>>().with_name("amount"),
        ];
    }

    unsafe impl<'a> StructEncodeVisitor for Coin<'a> {
        fn encode_field<E: Encoder>(&self, index: usize, encoder: &mut E) -> Result<(), EncodeError> {
            match index {
                0 => <&'a str as Value<'a>>::encode(&self.denom, encoder),
                1 => <u128 as Value<'a>>::encode(&self.amount, encoder),
                _ => Err(EncodeError::UnknownError),
            }
        }
    }

    impl<'a> Value<'a> for Coin<'a> {
        type Type = StructT<Coin<'a>>;
        type DecodeState = (<&'a str as Value<'a>>::DecodeState, <u128 as Value<'a>>::DecodeState);

        fn visit_decode_state<D: Decoder<'a>>(state: &mut Self::DecodeState, decoder: &mut D) -> Result<(), DecodeError> {
            struct Visitor<'b, 'a: 'b> {
                state: &'b mut <Coin<'a> as Value<'a>>::DecodeState,
            }
            unsafe impl<'b, 'a: 'b> StructSchema for Visitor<'b, 'a> {
                const FIELDS: &'static [Field<'static>] = Coin::<'a>::FIELDS;
            }
            unsafe impl<'b, 'a: 'b> StructDecodeVisitor<'a> for Visitor<'b, 'a> {
                fn decode_field<D: Decoder<'a>>(&mut self, index: usize, decoder: &mut D) -> Result<(), DecodeError> {
                    match index {
                        0 => <&'a str as Value<'a>>::visit_decode_state(&mut self.state.0, decoder),
                        1 => <u128 as Value<'a>>::visit_decode_state(&mut self.state.1, decoder),
                        _ => Err(DecodeError::UnknownFieldNumber),
                    }
                }
            }
            decoder.decode_struct(&mut Visitor { state })
        }

        fn finish_decode_state(state: Self::DecodeState, mem: &MemoryManager<'a, 'a>) -> Result<Self, DecodeError> {
            let states = (
                <&'a str as Value<'a>>::finish_decode_state(state.0, mem)?,
                <u128 as Value<'a>>::finish_decode_state(state.1, mem)?,
            );
            Ok(Coin { denom: states.0, amount: states.1 })
        }

        /// Encode the value to the encoder.
        fn encode<E: Encoder>(&self, encoder: &mut E) -> Result<(), EncodeError> {
            encoder.encode_struct(self)
        }
    }

    #[test]
    fn test_coin() {
        let coin = Coin {
            denom: "uatom",
            amount: 1234567890,
        };
        let mut bump = Bump::new();
        let scope = bump.as_scope();
        let mut mem = MemoryManager::new(scope);
        let res = encode_value(&coin, scope).unwrap();
        let decoded = decode_value::<Coin>(res, &mem).unwrap();
        assert_eq!(decoded, coin);
    }

    #[test]
    fn test_coins() {
        let coins = vec![Coin {
            denom: "uatom",
            amount: 1234567890,
        }, Coin {
            denom: "foo",
            amount: 9876543210,
        }];
        let mut bump = Bump::new();
        let res = encode_value(&coins, bump.as_scope()).unwrap();
        let mut mem = MemoryManager::new(bump.as_scope());
        let decoded = decode_value::<&[Coin]>(&res, &mut mem).unwrap();
        assert_eq!(decoded, coins);
    }
}
