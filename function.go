package nw

import (
	"encoding/json"

	"github.com/tidwall/gjson"
)

func covertRequestData(reqestData any) []byte {
	var b []byte
	switch tmp := reqestData.(type) {
	case string:
		b = []byte(tmp)
	case *string:
		b = []byte(*tmp)
	case gjson.Result:
		b = []byte(tmp.Raw)
	case *gjson.Result:
		b = []byte(tmp.Raw)
	case nil:
		b = []byte("")
	default:
		b, _ = json.Marshal(tmp)
	}
	return b
}
