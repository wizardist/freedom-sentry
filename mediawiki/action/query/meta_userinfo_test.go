package query

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestUserinfoMetaQuery_setResponse(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    Userinfo
		wantErr bool
	}{
		{
			name:    "Empty",
			json:    `{"userinfo":{"id":42,"name":"Bobby Tables"}}`,
			want:    Userinfo{Id: 42, Name: "Bobby Tables"},
			wantErr: false,
		},
		{
			name:    "One right",
			json:    `{"userinfo":{"id":42,"name":"Bobby Tables","rights":["test1"]}}`,
			want:    Userinfo{Id: 42, Name: "Bobby Tables", Rights: []string{"test1"}},
			wantErr: false,
		},
		{
			name:    "Multiple rights",
			json:    `{"userinfo":{"id":42,"name":"Bobby Tables","rights":["test1","test2"]}}`,
			want:    Userinfo{Id: 42, Name: "Bobby Tables", Rights: []string{"test1", "test2"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserinfoMetaQuery{}

			var payload map[string]interface{}
			err := json.Unmarshal([]byte(tt.json), &payload)
			if err != nil {
				t.Fatal(err)
			}

			err = u.setResponse(payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("setResponse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(u.GetUserinfo(), tt.want) {
				t.Errorf("GetUserinfo() got %v, wanted %v", u.GetUserinfo(), tt.want)
			}
		})
	}
}
