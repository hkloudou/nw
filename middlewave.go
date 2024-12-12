package nw

import (
	"fmt"
	"net/http"
)

type Middlewaves[T any] struct {
	reqHandlers   []func(req *http.Request)
	decodeHandler func(response *http.Response) *Result[T]
	resHandlers   []func(res *Result[T])
}

// UseRequest 添加请求预处理器
func (m *Middlewaves[T]) UseRequest(cb func(req *http.Request)) *Middlewaves[T] {
	m.reqHandlers = append(m.reqHandlers, cb)
	return m
}

// UseDecode 设置解码器
func (m *Middlewaves[T]) UseDecode(fun func(response *http.Response) *Result[T]) *Middlewaves[T] {
	m.decodeHandler = fun
	return m
}

// UseResponse 添加响应处理器
func (m *Middlewaves[T]) UseResponse(cb func(*Result[T])) *Middlewaves[T] {
	m.resHandlers = append(m.resHandlers, cb)
	return m
}

func (m *Middlewaves[T]) Copy() *Middlewaves[T] {
	ret := &Middlewaves[T]{
		reqHandlers: make([]func(req *http.Request), 0),
		decodeHandler: func(response *http.Response) *Result[T] {
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
func NewMiddlewaves[T any]() *Middlewaves[T] {
	return &Middlewaves[T]{
		reqHandlers: make([]func(req *http.Request), 0),
		decodeHandler: func(response *http.Response) *Result[T] {
			return WrapParseError[T](fmt.Errorf("no decode handler set"))
		},
		resHandlers: make([]func(res *Result[T]), 0),
	}
}
