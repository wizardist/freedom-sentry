package query

type Property interface {
	ToPropertyPayload() map[string]interface{}

	setResponse(json map[string]interface{}) error
}

type Meta interface {
	ToMetaPayload() map[string]interface{}

	setResponse(json map[string]interface{}) error
}
