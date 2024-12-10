package nw

// func PostJsonSteam[T any](opts ...NwOption) *res[T] {
// 	o := getDefaultOption(opts...)

// 	req, err := http.NewRequest("POST", o.site, o.postReader)
// 	if err != nil {
// 		return &res[T]{
// 			Code: 400,
// 			Msg:  err.Error(),
// 			Data: nil,
// 		}
// 	}

// 	fill(o, req)
// 	resp, err := o.client.Do(req)
// 	if err != nil {
// 		return &res[T]{
// 			Code: 400,
// 			Msg:  err.Error(),
// 			Data: nil,
// 		}
// 	}
// 	defer resp.Body.Close()

// 	return returnStream[T](resp.Body, o)
// }

// func PostJsonData[T any](data interface{}, opts ...NwOption) *res[T] {
// 	b, err := json.Marshal(data)
// 	if err != nil {
// 		return &res[T]{
// 			Code: 400,
// 			Msg:  err.Error(),
// 			Data: nil,
// 		}
// 	}
// 	opts = append(opts, WithPostData(bytes.NewReader(b)))
// 	return PostJsonSteam[T](opts...)
// }

// func fill(o *nwOption, req *http.Request) {
// 	req.Header.Set("Content-Type", "application/json")
// 	if o.header != nil {
// 		req.Header = o.header
// 	}

// 	for i := 0; i < len(o.mid.reqs); i++ {
// 		o.mid.reqs[i](req)
// 	}
// }

// type res[T any] struct {
// 	Code int
// 	Msg  string
// 	Data *T
// }

// func GetJsonData[T any](opts ...NwOption) *res[T] {
// 	o := getDefaultOption(opts...)
// 	req, err := http.NewRequest("GET", o.site, nil)
// 	if err != nil {
// 		return &res[T]{
// 			Code: 400,
// 			Msg:  err.Error(),
// 			Data: nil,
// 		}
// 	}
// 	fill(o, req)

// 	resp, err := o.client.Do(req)
// 	if err != nil {
// 		return &res[T]{
// 			Code: 400,
// 			Msg:  err.Error(),
// 			Data: nil,
// 		}
// 	}
// 	defer resp.Body.Close()

// 	for i := 0; i < len(o.mid.ress); i++ {
// 		o.mid.ress[i](resp)
// 	}

// 	if resp != nil && resp.StatusCode != 200 {
// 		return &res[T]{
// 			Code: 400,
// 			Msg:  fmt.Sprintf("err response code:%d", resp.StatusCode),
// 			Data: nil,
// 		}
// 		// return nil, fmt.Errorf("err response code:%d", resp.StatusCode)
// 	}
// 	return returnStream[T](resp.Body, o)
// }

// func returnStream[T any](stream io.ReadCloser, o *nwOption) *res[T] {
// 	body, err := io.ReadAll(stream)
// 	if err != nil {
// 		return &res[T]{
// 			Code: 400,
// 			Msg:  err.Error(),
// 			Data: nil,
// 		}
// 	}
// 	if o.log {
// 		fmt.Println("resposne", o.site, string(body))
// 	}

// 	g := gjson.ParseBytes(body)
// 	sg := func(keys ...string) gjson.Result {
// 		for i := 0; i < len(keys); i++ {
// 			if keys[i] == "" {
// 				if g.Exists() {
// 					return g
// 				}
// 			}
// 			r := g.Get(keys[i])
// 			if r.Exists() {
// 				return r
// 			}
// 			if i == len(keys)-1 {
// 				return r
// 			}
// 		}
// 		panic("")
// 	}
// 	// if sg(o.codeKeys...).Int() != 0 {
// 	// 	msg := sg(o.msgKeys...).String()
// 	// 	if len(msg) > 0 {
// 	// 		return nil, fmt.Errorf(msg)
// 	// 	}
// 	// 	return nil, fmt.Errorf("error fmt")
// 	// }

// 	code := sg(o.codeKeys...).Int()
// 	msg := sg(o.msgKeys...).String()

// 	var dataRaw = sg(o.dataKeys...).Raw
// 	var obj T
// 	if reflect.TypeOf(obj).String() == "gjson.Result" {
// 		if result, ok := interface{}(gjson.Parse(dataRaw)).(T); ok {
// 			return &res[T]{
// 				Code: int(code),
// 				Msg:  msg,
// 				Data: &result,
// 			}
// 		}
// 		return &res[T]{
// 			Code: 500,
// 			Msg:  "error fmt",
// 			Data: nil,
// 		}
// 	}

// 	if reflect.TypeOf(obj).Kind() == reflect.Slice {
// 		reflect.ValueOf(&obj).Elem().Set(reflect.MakeSlice(reflect.TypeOf(obj), 0, 0))
// 	}

// 	err = json.Unmarshal([]byte(dataRaw), &obj)
// 	if err != nil {
// 		return &res[T]{
// 			Code: 500,
// 			Msg:  "error fmt",
// 			Data: nil,
// 		}
// 	}
// 	return &res[T]{
// 		Code: int(code),
// 		Msg:  msg,
// 		Data: &obj,
// 	}
// }

// func returnJson[T any](stream io.ReadCloser, o *nwOption) *Result[T] {
// 	body, err := io.ReadAll(stream)
// 	if err != nil {
// 		// return w
// 		return &res[T]{Code: 400, Msg: err.Error()}
// 	}

// 	if o.log {
// 		fmt.Println("response", o.site, string(body))
// 	}

// 	var obj T
// 	switch reflect.TypeOf(obj).Kind() {
// 	case reflect.Struct, reflect.Slice:
// 		if err := json.Unmarshal(body, &obj); err != nil {
// 			return &res[T]{Code: 500, Msg: "JSON unmarshal failed: " + err.Error()}
// 		}
// 	default:
// 		return &res[T]{Code: 500, Msg: "Unsupported type"}
// 	}

// 	return &res[T]{Code: 200, Msg: "", Data: &obj}
// }
