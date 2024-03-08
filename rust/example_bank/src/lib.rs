pub mod example {
    pub mod bank {
        pub mod v1 {
            pub mod bank {
                include!("example/bank/v1/bank.rs");
            }
        }
    }
}

pub struct Bank {}
