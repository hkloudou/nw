package nw

import (
	"encoding/json"
	"io"
	"net/http"
)

func TestMain() {
	// nw.Then[]()
	Get("https://www.google.com/", NewMiddlewaves[string]().UseDecode(func(response *http.Response) *Result[string] {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return WrapParseError[string](err)
		}
		var obj string
		err = json.Unmarshal(body, &obj)
		if err != nil {
			return WrapParseError[string](err)
		}
		return WrapData(&obj)
	}))
}
