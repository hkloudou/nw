package nw

import (
	"encoding/json"
)

func TestMain() {
	// nw.Then[]()
	Get("https://www.google.com/", NewMiddlewaves[string]().UseDecode(func(data []byte) *Result[string] {
		var obj string
		err := json.Unmarshal(data, &obj)
		if err != nil {
			return WrapParseError[string](err)
		}
		return WrapData(&obj)
	}))
}
