use crate::decoder::{DecodeError, Decoder};
use crate::encoder::{EncodeError, Encoder};
use crate::field::Field;

/// StructCodec is the trait that should be derived to encode and decode a struct.
///
/// It should generally be used in conjunction with the `#[derive(StructCodec)]` attribute
/// attached to a struct definition.
/// It is unsafe to implement this trait manually, because the compiler cannot
///  guarantee correct implementations.
///
/// Any struct which contains fields which implement [`value::MaybeBorrowed`] can be derived
/// to implement this trait.
/// Structs and their fields may optionally contain a single lifetime parameter, in which
/// case decoded values will be borrowed from the input data wherever possible.
///
/// Example:
/// ```
/// #[derive(StructCodec)]
/// pub struct MyStruct<'a> {
///   pub field1: u8,
///   pub field2: &'a str,
/// }
///
///
/// #[derive(StructCodec)]
/// pub struct MyStruct2 {
///   pub field1: simple_time::Time,
///   pub field2: interchain_message_api::Address,
/// }
/// ```
pub unsafe trait StructCodec {
    /// A dummy function for derived macro type checking.
    fn dummy(&self);
}

pub unsafe trait StructSchema {
    const FIELDS: &'static [Field<'static>];
}

pub unsafe trait StructDecodeVisitor<'a>: StructSchema {
    fn decode_field<D: Decoder<'a>>(&mut self, index: usize, decoder: &mut D) -> Result<(), DecodeError>;
}

pub unsafe trait StructEncodeVisitor: StructSchema {
    fn encode_field<E: Encoder>(&self, index: usize, encoder: &mut E) -> Result<(), EncodeError>;
}

#[cfg(test)]
mod tests {
    use crate::decoder::{DecodeError, Decoder};
    use crate::encoder::{EncodeError, Encoder};
    use crate::field::Field;
    use crate::r#struct::{StructDecodeVisitor, StructEncodeVisitor, StructSchema};
    use crate::types::{to_field, StrT, StructT, UIntNT};
    use crate::value::ArgValue;

    struct Coin<'a> {
        denom: &'a str,
        amount: u128,
    }

    impl<'a> ArgValue<'a> for Coin<'a> {
        type Type = StructT<Coin<'a>>;
        type DecodeState = (<&'a str as ArgValue<'a>>::DecodeState, <u128 as ArgValue<'a>>::DecodeState);
        type MemoryHandle = (Option<<&'a str as ArgValue<'a>>::MemoryHandle>, Option<<u128 as ArgValue<'a>>::MemoryHandle>);

        fn visit_decode_state<D: Decoder<'a>>(state: &mut Self::DecodeState, decoder: &mut D) -> Result<(), DecodeError> {
            struct Visitor<'b, 'a:'b> {
                state: &'b mut <crate::r#struct::tests::Coin<'a> as ArgValue<'a>>::DecodeState,
            }
            unsafe impl <'b, 'a:'b> StructSchema for Visitor<'b, 'a> {
                const FIELDS: &'static [Field<'static>] = &[
                    to_field::<StrT>().with_name("denom"),
                    to_field::<UIntNT<16>>().with_name("amount"),
                ];
            }
            unsafe impl<'b, 'a:'b> StructDecodeVisitor<'a> for Visitor<'b, 'a> {
                fn decode_field<D: Decoder<'a>>(&mut self, index: usize, decoder: &mut D) -> Result<(), DecodeError> {
                    match index {
                        // 0 => <&'a str as ArgValue<'a>>::visit_decode_state(&mut self.state.0, decoder),
                        // 1 => <u128 as ArgValue<'a>>::visit_decode_state(&mut self.state.1, decoder),
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
            struct Visitor<'a> {
                state: &'a Coin<'a>,
            }
            unsafe impl <'a> StructSchema for Visitor<'a> {
                const FIELDS: &'static [Field<'static>] = &[
                    to_field::<StrT>().with_name("denom"),
                    to_field::<UIntNT<16>>().with_name("amount"),
                ];
            }
            unsafe impl<'a> StructEncodeVisitor for Visitor<'a> {
                fn encode_field<E: Encoder>(&self, index: usize, encoder: &mut E) -> Result<(), EncodeError> {
                    match index {
                        0 => <&'a str as ArgValue<'a>>::encode(&self.state.denom, encoder),
                        1 => <u128 as ArgValue<'a>>::encode(&self.state.amount, encoder),
                        _ => Err(EncodeError::UnknownError),
                    }
                }
            }

            encoder.encode_struct(&Visitor { state: self })
        }
    }

    #[test]
    fn test_coin() {
        let coin = Coin {
            denom: "uatom",
            amount: 1234567890,
        };
        // // let encoded = coin.encode_to_vec().unwrap();
        // // let decoded = Coin::decode_from_slice(&encoded).unwrap();
        // assert_eq!(coin, decoded);
    }
}