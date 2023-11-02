pub unsafe trait ZeroCopy {}

unsafe impl ZeroCopy for bool {}

unsafe impl ZeroCopy for rend::i32_le {}

unsafe impl ZeroCopy for rend::u32_le {}

unsafe impl ZeroCopy for rend::i64_le {}

unsafe impl ZeroCopy for rend::u64_le {}

