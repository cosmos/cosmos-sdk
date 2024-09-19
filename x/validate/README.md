# x/validate

The `x/validate` is an app module solely there to setup ante/post handlers on an runtime app (via baseapp options) and the tx validator on the runtime/v2 app (via app module). 

Module specific tx validators should be registered on their own modules.
