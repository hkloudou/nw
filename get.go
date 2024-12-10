package nw

import (
	"fmt"
	"io"
	"net/http"
)

func Get[T any](site string, mw *middlewaves[T], opts ...NwOption) *Result[T] {
	o := getDefaultOption(opts...)
	request, err := http.NewRequest("GET", site, nil)
	if err != nil {
		return WrapNetworkError[T](err)
	}

	// 应用请求中间件
	for _, handler := range mw.reqHandlers {
		handler(request)
	}

	// 执行请求
	response, err := o.client.Do(request)
	if err != nil {
		return WrapNetworkError[T](err)
	}
	defer response.Body.Close()

	// 检查 HTTP 状态码
	if response.StatusCode != http.StatusOK {
		return WrapApiError[T](response.StatusCode, fmt.Sprintf("unexpected status code: %d", response.StatusCode))
	}

	// 读取响应体
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return WrapParseError[T](err)
	}

	// 解码响应体
	result := mw.decodeHandler(body)

	// 应用响应中间件
	for _, handler := range mw.resHandlers {
		handler(result)
	}

	return result
}

// type NwClient struct {
// 	opt *nwOption
// }

// func Client(opts ...NwOption) *NwClient {
// 	o := getDefaultOption(opts...)
// 	return &NwClient{
// 		opt: o,
// 	}
// }

// func Get[T any](site string, mw *middlewaves, opts ...NwOption) *Result[T] {
// 	o := getDefaultOption(opts...)
// 	request, err := http.NewRequest("GET", site, nil)
// 	if err != nil {
// 		return WrapNetworkError[T](err)
// 	}

// 	// 应用请求中间件
// 	for _, handler := range mw.reqHandlers {
// 		handler(request)
// 	}

// 	// 执行请求
// 	response, err := o.client.Do(request)
// 	if err != nil {
// 		return WrapNetworkError[T](err)
// 	}
// 	defer response.Body.Close()

// 	// 检查 HTTP 状态码
// 	if response.StatusCode != http.StatusOK {
// 		return WrapApiError[T](response.StatusCode, fmt.Sprintf("unexpected status code: %d", response.StatusCode))
// 	}

// 	// 读取响应体
// 	body, err := io.ReadAll(response.Body)
// 	if err != nil {
// 		return WrapParseError[T](err)
// 	}

// 	// 解码响应体
// 	result := mw.decodeHandler(body)

// 	// 应用响应中间件
// 	for _, handler := range mw.resHandlers {
// 		handler(result)
// 	}

// 	if result.err != nil {
// 		var obj T
// 		return WrapResult(&obj, result.err)
// 	}

// 	if d, ok := result.Data.(*T); ok {
// 		return WrapData(d)
// 	}

// 	return result
// }

// func GetJson[T any](nw *NwClient, site string, prepare func(req *http.Request)) *Result[T] {
// 	return Get[T](nw, site, func(req *http.Request) {
// 		jsonPrepare(req)
// 		if prepare != nil {
// 			prepare(req)
// 		}
// 	}, jsonDecode)
// }

// func jsonPrepare(req *http.Request) {
// 	// 设置常见的 JSON 请求头
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Accept", "application/json")
// 	// 添加其他自定义头
// 	// req.Header.Set("Authorization", "Bearer your-token")
// }

// func jsonDecode[T any](data []byte) *Result[T] {
// 	var obj T
// 	objType := reflect.TypeOf(obj)

// 	// 特殊处理 gjson.Result 类型
// 	if objType.String() == "gjson.Result" {
// 		result := gjson.ParseBytes(data)
// 		if res, ok := interface{}(result).(T); ok {
// 			return WrapData(&res)
// 		}
// 		return WrapParseError[T](fmt.Errorf("failed to decode into gjson.Result"))
// 	}

// 	// 如果 obj 是切片类型，初始化空切片
// 	if objType.Kind() == reflect.Slice {
// 		objValue := reflect.MakeSlice(objType, 0, 0)
// 		obj = objValue.Interface().(T)
// 	}

// 	// 尝试 JSON 反序列化
// 	err := json.Unmarshal(data, &obj)
// 	if err != nil {
// 		return WrapParseError[T](err)
// 	}

// 	return WrapData(&obj)
// }
