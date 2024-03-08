#[repr(C)]
pub struct ModuleBundleDescriptor {
    pub modules: ::zeropb::Repeated<crate::cosmos::core::v1alpha1::ModuleInitDescriptor>,
}
unsafe impl zeropb::ZeroCopy for ModuleBundleDescriptor {}
#[repr(C)]
pub struct ModuleInitDescriptor {
    pub module_name: ::zeropb::Str,
    pub inputs: ::zeropb::Repeated<crate::cosmos::core::v1alpha1::Input>,
    pub outputs: ::zeropb::Repeated<crate::cosmos::core::v1alpha1::ModuleOutput>,
}
unsafe impl zeropb::ZeroCopy for ModuleInitDescriptor {}
#[repr(C)]
pub struct Input {
    pub input: crate::cosmos::core::v1alpha1::Input::InputType,
    pub optional: bool,
}
unsafe impl zeropb::ZeroCopy for Input {}
#[repr(C)]
pub struct ModuleOutput {
    pub service: crate::cosmos::core::v1alpha1::ProtoService,
    pub message_handler: crate::cosmos::core::v1alpha1::ProtoMessageHandler,
    pub store: crate::cosmos::core::v1alpha1::StoreService,
    pub event_hook: crate::cosmos::core::v1alpha1::ProtoMessageHandler,
    pub pre_handler: crate::cosmos::core::v1alpha1::ProtoMessageHandler,
    pub post_handler: crate::cosmos::core::v1alpha1::ProtoMessageHandler,
}
unsafe impl zeropb::ZeroCopy for ModuleOutput {}
#[repr(C)]
pub struct ProtoService {
    pub name: ::zeropb::Str,
}
unsafe impl zeropb::ZeroCopy for ProtoService {}
#[repr(C)]
pub struct ProtoMessageHandler {
    pub name: ::zeropb::Str,
    pub response_name: ::zeropb::Str,
    pub virtual_file_descriptor: ::zeropb::Bytes,
}
unsafe impl zeropb::ZeroCopy for ProtoMessageHandler {}
#[repr(C)]
pub struct StoreService {}
unsafe impl zeropb::ZeroCopy for StoreService {}
#[repr(C)]
pub struct DynamicProtoClient {}
unsafe impl zeropb::ZeroCopy for DynamicProtoClient {}
#[repr(C)]
pub struct AccountHandlerDescriptor {
    pub handler_name: ::zeropb::Str,
    pub inputs: ::zeropb::Repeated<crate::cosmos::core::v1alpha1::Input>,
    pub outputs: ::zeropb::Repeated<crate::cosmos::core::v1alpha1::AccountOutput>,
}
unsafe impl zeropb::ZeroCopy for AccountHandlerDescriptor {}
#[repr(C)]
pub struct AccountOutput {
    pub service: crate::cosmos::core::v1alpha1::ProtoService,
    pub message_handler: crate::cosmos::core::v1alpha1::ProtoMessageHandler,
}
unsafe impl zeropb::ZeroCopy for AccountOutput {}
