use std::borrow::Cow;
use arrayvec::ArrayString;

struct MsgSend {
    // all this is heap allocated
    from: String,
    to: String,
    denom: String,
    amount: String,
}

struct MsgSendFixed {
    // all this stuff is stack allocated
    from: ArrayString<64>,
    to: ArrayString<64>,
    denom: ArrayString<32>,
    amount: u128,
}

struct MsgSendBorrowed<'a> {
    from: Cow<'a, str>,  // borrowed or heap allocated
    to: Cow<'a, str>,
    denom: Cow<'a, str>,
    amount: u128,
}

fn send(from: &str, to: &str, denom: &str, amount: &u128)  {
}

fn example_call_send() {
    let from = ArrayString::from("from");
    let to = ArrayString::from("to");
    let denom = ArrayString::from("denom");
    let amount = 100u128;
    send(&from, &to, &denom, &amount);
}