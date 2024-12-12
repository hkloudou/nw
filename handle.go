package nw

import (
	"fmt"
	"io"
	"net/http"
)

func handleRequest[T any](site string, o *nwOption, request *http.Request, mw *Middlewaves[T]) *Result[T] {
	if o.log {
		fmt.Printf("\u001b[44m\u001b[37m%s \u001b[0m %s\n", request.Method, site)
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
	site = response.Request.URL.String()
	if o.log {
		fmt.Println("\u001b[42m\u001b[37m ↓ \u001b[0m ", site, "\n", string(body))
	}

	// 解码响应体
	result := mw.decodeHandler(body)

	// 应用响应中间件
	for _, handler := range mw.resHandlers {
		if result.IsError() {
			return result
		}
		handler(result)
	}

	return result
}
