package nw

import (
	"fmt"
	"net/http"
)

type middlewaves[T any] struct {
	reqHandlers   []func(req *http.Request)
	decodeHandler func(data []byte) *Result[T]
	resHandlers   []func(res *Result[T])
}

// UseRequest 添加请求预处理器
func (m *middlewaves[T]) UseRequest(cb func(req *http.Request)) *middlewaves[T] {
	m.reqHandlers = append(m.reqHandlers, cb)
	return m
}

// UseDecode 设置解码器
func (m *middlewaves[T]) UseDecode(fun func(data []byte) *Result[T]) *middlewaves[T] {
	m.decodeHandler = fun
	return m
}

// UseResponse 添加响应处理器
func (m *middlewaves[T]) UseResponse(cb func(*Result[T])) *middlewaves[T] {
	m.resHandlers = append(m.resHandlers, cb)
	return m
}

func (m *middlewaves[T]) Copy() *middlewaves[T] {
	ret := &middlewaves[T]{
		reqHandlers: make([]func(req *http.Request), 0),
		decodeHandler: func(data []byte) *Result[T] {
			return WrapParseError[T](fmt.Errorf("no decode handler set"))
		},
		resHandlers: make([]func(res *Result[T]), 0),
	}

	ret.reqHandlers = append(ret.reqHandlers, m.reqHandlers...)
	ret.resHandlers = append(ret.resHandlers, m.resHandlers...)
	ret.decodeHandler = m.decodeHandler
	return ret
}

// newMiddlewaves 创建新的 middlewaves 实例，仅包内可用
func NewMiddlewaves[T any]() *middlewaves[T] {
	return &middlewaves[T]{
		reqHandlers: make([]func(req *http.Request), 0),
		decodeHandler: func(data []byte) *Result[T] {
			return WrapParseError[T](fmt.Errorf("no decode handler set"))
		},
		resHandlers: make([]func(res *Result[T]), 0),
	}
}
