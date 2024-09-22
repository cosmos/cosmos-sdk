use bump_scope::Bump;

pub struct Input<'a> {
    pub input: &'a [u8],
    pub bump_scope: &'a bump_scope::BumpScope<'a>,
}

pub trait Deserializer<'a> {}

pub trait Visitor<'a> {}

#[cfg(test)]
mod tests {
    use alloc::boxed::Box;
    use alloc::string::String;
    use alloc::vec;
    use core::any::Any;
    use bump_scope::{bump_vec, mut_bump_vec, Bump, BumpBox, BumpScope, BumpVec, MutBumpVec};
    use super::*;
    extern crate std;

    struct HasString {
        s: std::string::String,
    }

    struct DoDrop {}
    impl Drop for DoDrop {
        fn drop(&mut self) {
            std::println!("do drop");
        }
    }

    fn test1<'a: 'b, 'b>(scope: &'b BumpScope<'a>) -> (Box<dyn Drop + 'b>, &'a [HasString]) {
        struct Dropper<'a> {
            str_box: BumpBox<'a, [HasString]>,
            do_drop: DoDrop,
        }
        impl Drop for Dropper<'_> {
            fn drop(&mut self) {
                std::println!("dropped");
            }
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
            let dropper = Dropper {
                str_box,
                do_drop: DoDrop {},
            };

            (Box::new(dropper), &*str_slice)
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
        drop(todrop);
    }
}