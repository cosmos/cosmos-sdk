#[non_exhaustive]
pub struct MessageDescriptor<'a> {
    pub request_type: &'a str,
    pub response_type: &'a str,
    pub error_type: &'a str,
    pub events: &'a [&'a str],
}