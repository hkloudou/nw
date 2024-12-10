package nw

import (
	"fmt"
)

// Result 泛型结构体，用于封装 API 响应结果
type Result[T any] struct {
	err  error // 所有错误统一存储
	Data *T    // 成功时的结果数据
}

// IsError 检查是否有错误
func (r *Result[T]) IsError() bool {
	return r.err != nil
}

func (r *Result[T]) IsNetworkError() bool {
	_, ok := r.err.(*NetworkError)
	return ok
}

func (r *Result[T]) IsParseError() bool {
	_, ok := r.err.(*ParseError)
	return ok
}

func (r *Result[T]) IsApiError() bool {
	_, ok := r.err.(*ApiError)
	return ok
}

// GetData 获取数据，如果存在错误则返回 nil
func (r *Result[T]) GetData() *T {
	if r.IsError() {
		return nil
	}
	return r.Data
}

// GetError 获取错误
func (r *Result[T]) GetError() error {
	return r.err
}

// Then 方法调用外部的泛型函数实现链式调用
// func (r *Result[T]) Then(fn func(data *T) *Result[any]) *Result[any] {
// 	if r.IsError() || r.Data == nil {
// 		// 如果有错误，直接将错误传递给新的 Result
// 		return &Result[any]{err: r.err}
// 	}
// 	return fn(r.Data)
// }

// Catch 错误捕获处理
func (r *Result[T]) Catch(fn func(err error)) *Result[T] {
	if r.IsError() {
		fn(r.err) // 调用错误处理函数
	}
	return r // 返回当前 Result，以支持链式调用
}

// WrapResult 创建一个包含错误和数据的 Result
func WrapResult[T any](data *T, err error) *Result[T] {
	return &Result[T]{
		err:  err,
		Data: data,
	}
}

// 自定义错误类型

// NetworkError 表示网络相关的错误
type NetworkError struct {
	Inner error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %v", e.Inner)
}

// ParseError 表示 JSON 解析相关的错误
type ParseError struct {
	Inner error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error: %v", e.Inner)
}

// ApiError 表示 API 返回的业务错误
type ApiError struct {
	Code int
	Msg  string
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("api error: code=%d, msg=%s", e.Code, e.Msg)
}

// WrapNetworkError 包装网络错误
func WrapNetworkError[T any](err error) *Result[T] {
	return WrapResult[T](nil, &NetworkError{Inner: err})
}

// WrapParseError 包装 JSON 解析错误
func WrapParseError[T any](err error) *Result[T] {
	return WrapResult[T](nil, &ParseError{Inner: err})
}

// WrapApiError 包装 API 错误
func WrapApiError[T any](code int, msg string) *Result[T] {
	return WrapResult[T](nil, &ApiError{Code: code, Msg: msg})
}

// WrapData 包装成功结果
func WrapData[T any](data *T) *Result[T] {
	return WrapResult[T](data, nil)
}

// Then 链式处理并返回不同类型的结果
// func Then[T, U any](r *Result[T], fn func(data *T) *Result[U]) *Result[U] {
// 	if r.IsError() || r.Data == nil {
// 		// 如果有错误，直接将错误传递给新的 Result
// 		return &Result[U]{err: r.err}
// 	}
// 	return fn(r.Data) // 成功时调用传入的处理函数
// }

// 顶层函数 Then
func Then[T, U any](r *Result[T], fn func(data *T) *Result[U]) *Result[U] {
	if r.IsError() {
		return &Result[U]{err: r.err}
	}
	return fn(r.Data)
}
