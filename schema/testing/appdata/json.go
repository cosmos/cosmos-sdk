package appdatatest

import (
	"pgregory.net/rapid"

	"cosmossdk.io/schema/appdata"
)

func jsonValueGen() *rapid.Generator[any] {
	return rapid.OneOf(
		rapid.Bool().AsAny(),
		rapid.Float64().AsAny(),
		rapid.String().AsAny(),
		rapid.MapOf(rapid.String(), rapid.Deferred(jsonValueGen)).AsAny(),
		rapid.SliceOf(rapid.Deferred(jsonValueGen)).AsAny(),
	)
}

var JSONValueGen = jsonValueGen()

var JSONObjectGen = rapid.MapOf(rapid.String(), JSONValueGen)

var JSONArrayGen = rapid.SliceOf(JSONValueGen)

func JSONObjectWithKeys(keys ...string) *rapid.Generator[map[string]interface{}] {
	return rapid.MapOf(rapid.SampledFrom(keys), JSONValueGen)
}

func StringMapWithKeys(keys ...string) *rapid.Generator[map[string]string] {
	return rapid.MapOf(rapid.SampledFrom(keys), rapid.String())
}

// events can consist of names separated by dots, e.g. "message.sent"
const eventTypeFormat = `^([a-zA-Z_][a-zA-Z0-9_]*\.)*[A-Za-z_][A-Za-z0-9_]$`

var DefaultEventDataGen = rapid.Custom(func(t *rapid.T) appdata.EventData {
	return appdata.EventData{
		Type: rapid.StringMatching(`^$`).Draw(t, "type"),
	}
})
