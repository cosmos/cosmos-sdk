package appdatatest

//var JSONObjectGen *rapid.Generator[map[string]any] = rapid.MapOf(rapid.String(), JSONValueGen)
//
//var JSONArrayGen *rapid.Generator[[]any] = rapid.SliceOf(JSONValueGen)
//
//var JSONValueGen *rapid.Generator[any] = rapid.OneOf(
//	rapid.Bool().AsAny(),
//	rapid.Float64().AsAny(),
//	rapid.String().AsAny(),
//	JSONObjectGen.AsAny(),
//	JSONArrayGen.AsAny(),
//)
//
//func JSONObjectWithKeys(keys ...string) *rapid.Generator[map[string]interface{}] {
//	return rapid.MapOf(rapid.SampledFrom(keys), JSONValueGen)
//}
//
//func StringMapWithKeys(keys ...string) *rapid.Generator[map[string]string] {
//	return rapid.MapOf(rapid.SampledFrom(keys), rapid.String())
//}
//
//// events can consist of names separated by dots, e.g. "message.sent"
//const eventTypeFormat = `^([a-zA-Z_][a-zA-Z0-9_]*\.)*[A-Za-z_][A-Za-z0-9_]$`
//
//var DefaultEventDataGen = rapid.Custom(func(t *rapid.T) appdata.EventData {
//	return appdata.EventData{
//		Type: rapid.StringMatching(`^$`).Draw(t, "type"),
//	}
//})
