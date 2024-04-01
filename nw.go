package nw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/tidwall/gjson"
)

func PostJsonSteam[T any](opts ...NwOption) (*T, error) {
	o := getDefaultOption(opts...)

	req, err := http.NewRequest("POST", o.site, o.postReader)
	if err != nil {
		return nil, err
	}

	fill(o, req)
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return returnStream[T](resp.Body, o)
}

func PostJsonData[T any](data interface{}, opts ...NwOption) (*T, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	opts = append(opts, WithPostData(bytes.NewReader(b)))
	return PostJsonSteam[T](opts...)
}

func fill(o *nwOption, req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if o.header != nil {
		req.Header = o.header
	}
}

func GetJsonData[T any](opts ...NwOption) (*T, error) {
	o := getDefaultOption(opts...)
	req, err := http.NewRequest("GET", o.site, nil)
	if err != nil {
		return nil, err
	}
	fill(o, req)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp != nil && resp.StatusCode != 200 {
		return nil, fmt.Errorf("err response code:%d", resp.StatusCode)
	}
	return returnStream[T](resp.Body, o)
}

func returnStream[T any](stream io.ReadCloser, o *nwOption) (*T, error) {
	body, err := io.ReadAll(stream)
	if err != nil {
		return nil, err
	}
	if o.log {
		fmt.Println("resposne", o.site, string(body))
	}
	g := gjson.ParseBytes(body)
	sg := func(keys ...string) gjson.Result {
		for i := 0; i < len(keys); i++ {
			if keys[i] == "" {
				if g.Exists() {
					return g
				}
			}
			r := g.Get(keys[i])
			if r.Exists() {
				return r
			}
			if i == len(keys)-1 {
				return r
			}
		}
		panic("")
	}
	if sg(o.codeKeys...).Int() != 0 {
		msg := sg(o.msgKeys...).String()
		if len(msg) > 0 {
			return nil, fmt.Errorf(msg)
		}
		return nil, fmt.Errorf("error fmt")
	}
	var dataRaw = sg(o.dataKeys...).Raw
	var obj T
	if reflect.TypeOf(obj).String() == "gjson.Result" {
		if result, ok := interface{}(gjson.Parse(dataRaw)).(T); ok {
			return &result, nil
		}
		return nil, fmt.Errorf("err fmt")
	}

	err = json.Unmarshal([]byte(dataRaw), &obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}
