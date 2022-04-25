package mediawiki

type Action interface {
	IsWriteAction() bool
	ToActionPayload() map[string]interface{}
	SetResponse(json map[string]interface{}) error
}
