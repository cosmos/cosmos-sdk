#[repr(C)]
pub struct Foo {
    pub a: u32,
    pub b: u16,
    pub c: u64,
}

#[no_mangle]
pub extern fn add(left: usize, right: *mut u8) -> usize {
    left + right.len()
}

#[no_mangle]
pub extern fn foo(bar: Bar) -> Foo {
    Foo{ a: 1, b: 2, c: 3 }
}

#[repr(u8)]
pub enum Bar {
    A,
    B,
    C,
}