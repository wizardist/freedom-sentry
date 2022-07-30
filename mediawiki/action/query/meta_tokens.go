package query

type TokensMetaQuery struct {
	// FIXME: Only supports CSRF token requests

	Type []string

	tokens struct {
		Csrf string
	}
}

func (qm TokensMetaQuery) ToMetaPayload() map[string]interface{} {
	return map[string]interface{}{
		"meta": "tokens",
		"type": qm.Type,
	}
}

func (qm TokensMetaQuery) GetTokens() struct{ Csrf string } {
	return qm.tokens
}

func (qm *TokensMetaQuery) setResponse(json map[string]interface{}) error {
	tokens := json["tokens"].(map[string]interface{})

	qm.tokens.Csrf = tokens["csrftoken"].(string)

	return nil
}
