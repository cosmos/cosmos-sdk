use alloc::string::String;
use bump_scope::{BumpScope, BumpString};
use crate::decoder::DecodeError;
use crate::list::ListVisitor;
use crate::r#struct::StructDecodeVisitor;
use crate::value::ArgValue;

pub fn decode_value<'b, 'a: 'b, V: ArgValue<'a>>(input: &'a [u8], scope: &'b BumpScope<'a>) -> Result<(V, Option<V::MemoryHandle>), DecodeError> {
    let mut decoder = Decoder { buf: input, scope };
    let mut decode_state = V::DecodeState::default();
    V::visit_decode_state(&mut decode_state, &mut decoder)?;
    V::finish_decode_state(decode_state)
}

struct Decoder<'b, 'a: 'b> {
    buf: &'a [u8],
    scope: &'b BumpScope<'a>,
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

    fn decode_borrowed_str(&mut self) -> Result<Result<&'a str, BumpString<'a, 'a>>, DecodeError> {
        let bz = self.buf;
        self.buf = &[];
        Ok(Ok(core::str::from_utf8(bz).map_err(|_| DecodeError::InvalidData)?))
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
        // let size = self.decode_u32()? as usize;
        // visitor.init(size, &mut self.scope)?;
        // let mut inner = InnerDecoder { outer: self };
        // for _ in 0..size {
        //     visitor.next(&mut inner)?;
        // }
        todo!();
        Ok(())
    }

    fn scope(&self) -> &'a BumpScope<'a> {
        // &self.scope
        todo!();
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

    fn decode_borrowed_str(&mut self) -> Result<Result<&'a str, BumpString<'a, 'a>>, DecodeError> {
        let size = self.decode_u32()? as usize;
        let bz = self.outer.read_bytes(size)?;
        Ok(Ok(core::str::from_utf8(bz).map_err(|_| DecodeError::InvalidData)?))
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

    fn scope(&self) -> &'a BumpScope<'a> {
        todo!()
    }
}

#[cfg(test)]
mod tests {
    extern crate std;
    use bump_scope::{Bump, BumpScope};
    use crate::binary::decoder::decode_value;
    use crate::binary::encoder::encode_value;
    use crate::decoder::{DecodeError, Decoder};
    use crate::encoder::{EncodeError, Encoder};
    use crate::field::Field;
    use crate::r#struct::{StructDecodeVisitor, StructEncodeVisitor, StructSchema};
    use crate::types::{to_field, StrT, StructT, UIntNT};
    use crate::value::ArgValue;

    #[test]
    fn test_u32_decode() {
        let buf: [u8; 4] = [10, 0, 0, 0];
        let mut bump = Bump::new();
        let (x, _) = decode_value::<u32>(&buf, &mut bump.as_mut_scope()).unwrap();
        assert_eq!(x, 10);
    }

    #[test]
    fn test_decode_borrowed_string() {
        let str = "hello";
        let mut bump = Bump::new();
        let (x, _) = decode_value::<&str>(str.as_bytes(), &mut bump.as_mut_scope()).unwrap();
        assert_eq!(x, "hello");
    }

    #[test]
    fn test_decode_owned_string() {
        let str = "hello";
        let mut bump = Bump::new();
        let (x, _) = decode_value::<alloc::string::String>(str.as_bytes(), &mut bump.as_mut_scope()).unwrap();
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
                0 => <&'a str as ArgValue<'a>>::encode(&self.denom, encoder),
                1 => <u128 as ArgValue<'a>>::encode(&self.amount, encoder),
                _ => Err(EncodeError::UnknownError),
            }
        }
    }

    impl<'a> ArgValue<'a> for Coin<'a> {
        type Type = StructT<Coin<'a>>;
        type DecodeState = (<&'a str as ArgValue<'a>>::DecodeState, <u128 as ArgValue<'a>>::DecodeState);
        type MemoryHandle = (Option<<&'a str as ArgValue<'a>>::MemoryHandle>, Option<<u128 as ArgValue<'a>>::MemoryHandle>);

        fn visit_decode_state<D: Decoder<'a>>(state: &mut Self::DecodeState, decoder: &mut D) -> Result<(), DecodeError> {
            struct Visitor<'b, 'a: 'b> {
                state: &'b mut <Coin<'a> as ArgValue<'a>>::DecodeState,
            }
            unsafe impl<'b, 'a: 'b> StructSchema for Visitor<'b, 'a> {
                const FIELDS: &'static [Field<'static>] = Coin::<'a>::FIELDS;
            }
            unsafe impl<'b, 'a: 'b> StructDecodeVisitor<'a> for Visitor<'b, 'a> {
                fn decode_field<D: Decoder<'a>>(&mut self, index: usize, decoder: &mut D) -> Result<(), DecodeError> {
                    match index {
                        0 => <&'a str as ArgValue<'a>>::visit_decode_state(&mut self.state.0, decoder),
                        1 => <u128 as ArgValue<'a>>::visit_decode_state(&mut self.state.1, decoder),
                        _ => Err(DecodeError::UnknownFieldNumber),
                    }
                }
            }
            decoder.decode_struct(&mut Visitor { state })
        }

        fn finish_decode_state(state: Self::DecodeState) -> Result<(Self, Option<Self::MemoryHandle>), DecodeError> {
            let states = (
                <&'a str as ArgValue<'a>>::finish_decode_state(state.0)?,
                <u128 as ArgValue<'a>>::finish_decode_state(state.1)?,
            );
            let mut mem = None;
            if states.0.1.is_some() || states.1.1.is_some() {
                mem = Some((states.0.1, states.1.1));
            }
            Ok((Coin { denom: states.0.0, amount: states.1.0 }, mem))
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
        let res = encode_value(&coin, bump.as_scope()).unwrap();
        let (decoded, mem) = decode_value::<Coin>(&res, bump.as_scope()).unwrap();
        assert_eq!(decoded, coin);
        let _ = mem;
    }
}
