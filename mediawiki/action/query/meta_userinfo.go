package query

type Userinfo struct {
	Id     uint64
	Name   string
	Rights []string
}

type UserinfoMetaQuery struct {
	// Which pieces of information to include, one of uiprop
	Properties []string

	userinfo Userinfo
}

func (u UserinfoMetaQuery) ToMetaPayload() map[string]interface{} {
	return map[string]interface{}{
		"meta":   "userinfo",
		"uiprop": u.Properties,
	}
}

func (u UserinfoMetaQuery) GetUserinfo() Userinfo {
	return u.userinfo
}

func (u *UserinfoMetaQuery) setResponse(json map[string]interface{}) error {
	userinfo := json["userinfo"].(map[string]interface{})

	u.userinfo.Id = uint64(userinfo["id"].(float64))
	u.userinfo.Name = userinfo["name"].(string)
	if v, ok := userinfo["rights"].([]interface{}); ok {
		u.userinfo.Rights = make([]string, len(v))
		for i, r := range v {
			u.userinfo.Rights[i] = r.(string)
		}
	}

	return nil
}
