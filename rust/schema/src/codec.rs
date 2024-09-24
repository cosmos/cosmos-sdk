use bump_scope::{Bump, BumpString, BumpVec};

pub struct Input<'a> {
    pub input: &'a [u8],
    pub bump_scope: &'a bump_scope::BumpScope<'a>,
}

pub trait Deserializer<'a> {}

pub trait Visitor<'a> {}

trait DeferDrop {}
impl<T> DeferDrop for T {}


#[cfg(test)]
mod tests {
    use alloc::boxed::Box;
    use alloc::string::String;
    use alloc::vec;
    use core::any::Any;
    use core::ptr::NonNull;
    use bump_scope::{bump_vec, mut_bump_vec, Bump, BumpBox, BumpScope, BumpVec, MutBumpVec};
    use super::*;
    extern crate std;

    struct HasString {
        s: std::string::String,
    }

    impl Drop for HasString {
        fn drop(&mut self) {
            std::println!("do drop {}", self.s);
        }
    }

    fn test1<'a: 'b, 'b>(scope: &'b BumpScope<'a>) -> (*mut (dyn DeferDrop + 'b), &'a [HasString]) {
        struct Dropper<'a> {
            str_box: BumpBox<'a, [HasString]>,
        }
        let mut strings = BumpVec::new_in(scope);
        strings.push(HasString {
            s: String::from("hello"),
        });
        strings.push(HasString {
            s: String::from("world"),
        });
        strings.push(HasString {
            s: String::from("foo"),
        });
        strings.push(HasString {
            s: String::from("bar"),
        });
        let str_box = strings.into_boxed_slice();
        unsafe {
            let str_slice = str_box.as_non_null_slice().as_ptr() as *const [HasString];
            let dropper = scope.alloc(Dropper {
                str_box,
            });
            (dropper.into_raw().as_ptr() as *mut (dyn DeferDrop + 'b), &*str_slice)
        }
    }

    #[test]
    fn test_test1() {
        let bump = Bump::new();
        let scope = bump.as_scope();
        let mut todrop = bump_vec![in scope];
        let (dropper, strs) = test1(&scope);
        todrop.push(dropper);
        for s in strs {
            std::println!("{}", s.s);
        }
        unsafe {
            for d in todrop.drain(..) {
                let _ = *d;
            }
        }
        drop(todrop);
    }

    fn test2() -> *const dyn DeferDrop {
        struct Foo {}
        impl Drop for Foo {
            fn drop(&mut self) {
                std::println!("dropping foo");
            }
        }
        let foo = Foo {};
        &foo as *const Foo as *const dyn DeferDrop
    }

    #[test]
    fn test_test2() {
        let _ = test2();
    }
}

trait Helper<'a> {
    type H<'b>
    where
        'a: 'b;
}

impl<'a> Helper<'a> for &'a str {
    type H<'b>
    where
        'a: 'b,
    = BumpString<'b, 'a>;
}


impl<'a, T> Helper<'a> for &'a [T] {
    type H<'b>
    where
        'a: 'b,
    = BumpVec<'b, 'a, T>;
}

trait Test1 {
    fn test1() -> impl DeferDrop;
}

struct Coin<'a> {
    denom: &'a str,
    amount: u128,
}

impl<'a> Test1 for Coin<'a> {
    fn test1() -> impl DeferDrop {
        // struct Dropper<'a, 'b> {
        //     denom: <&'a str as Helper<'a>>::H<'b>,
        //     amount: (),
        // }
        // Dropper {
        //     denom: todo!(),
        //     amount: (),
        // }
        todo!()
    }
}
