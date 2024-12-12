package nw

import (
	"encoding/json"
	"io"
	"net/http"
)

func TestMain() {
	// c2, _ := cookiejar.New(&cookiejar.Options{})
	// c2.SetCookies()
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
