// #[repr(C)]
// pub struct OneOf<T, const MaxValue: u32> {
//     value: T
// }
//
// impl <T, const MaxValue: u32> OneOf<T, MaxValue> {
//     fn get(&self) -> Result<&T, u32> {
//         let discriminant = unsafe { *<*const _>::from(self).cast::<u32>() };
//         if discriminant > MaxValue {
//             Err(discriminant)
//         } else {
//             Ok(&self.value)
//         }
//     }
//
//     fn set(&mut self, value: T) {
//         self.value = value
//     }
// }
//
// #[repr(C)]
// pub struct Option<T> {
//     some: bool,
//     value: T,
// }
