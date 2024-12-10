package nw

import (
	"fmt"
	"io"
	"net/http"
)

func handleRequest[T any](site string, o *nwOption, request *http.Request, mw *middlewaves[T]) *Result[T] {
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
