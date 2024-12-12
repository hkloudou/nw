package nw

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func Get[T any](site string, mw *Middlewaves[T], opts ...NwOption) *Result[T] {
	o := getDefaultOption(opts...)
	request, err := http.NewRequest("GET", site, nil)
	if err != nil {
		return WrapNetworkError[T](err)
	}

	return handleRequest(site, o, request, mw)
}

func PostJson[T any](site string, reqestData any, mw *Middlewaves[T], opts ...NwOption) *Result[T] {
	b, err := json.Marshal(reqestData)
	if err != nil {
		return WrapParseError[T](err)
	}
	if reqestData == nil {
		b = []byte("")
	}
	o := getDefaultOption(opts...)
	request, err := http.NewRequest("POST", site, bytes.NewReader(b))
	if err != nil {
		return WrapNetworkError[T](err)
	}

	return handleRequest(site, o, request, mw)
}
