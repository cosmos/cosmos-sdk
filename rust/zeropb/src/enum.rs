// #[repr(C)]
// pub struct Enum<T: Copy, const MaxValue: u8> {
//     value: u8,
//     _phantom: PhantomData<T>,
// }
//
// impl<T: Copy, const MaxValue: u8> Enum<T, MaxValue> {
//     fn get(&self) -> Result<T, u8> {
//         if self.value > MaxValue {
//             Err(self.value)
//         } else {
//             Ok(self.value as T)
//         }
//     }
//
//     fn set(&mut self, value: T) {
//         self.value = value
//     }
// }
//
