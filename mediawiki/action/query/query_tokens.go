package query

type TokensQueryMeta struct {
	// FIXME: Only supports CSRF token requests

	Type []string

	tokens struct {
		Csrf string
	}
}

func (qm TokensQueryMeta) ToMetaPayload() map[string]interface{} {
	return map[string]interface{}{
		"meta": "tokens",
		"type": qm.Type,
	}
}

func (qm TokensQueryMeta) GetTokens() struct{ Csrf string } {
	return qm.tokens
}

func (qm *TokensQueryMeta) setResponse(json map[string]interface{}) error {
	tokens := json["tokens"].(map[string]interface{})

	qm.tokens.Csrf = tokens["csrftoken"].(string)

	return nil
}
