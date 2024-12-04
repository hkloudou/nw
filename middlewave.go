package nw

import "net/http"

type middlewaves struct {
	reqs []func(req *http.Request)
	ress []func(res *http.Response)
}

func (m *middlewaves) UseRequest(cb func(req *http.Request)) {
	if len(m.reqs) == 0 {
		m.reqs = make([]func(req *http.Request), 0)
	}
	m.reqs = append(m.reqs, cb)
}

func (m *middlewaves) UseReponse(cb func(res *http.Response)) {
	if len(m.reqs) == 0 {
		m.ress = make([]func(req *http.Response), 0)
	}
	m.ress = append(m.ress, cb)
}

func NewMiddlewaves() *middlewaves {
	return &middlewaves{
		reqs: make([]func(req *http.Request), 0),
		ress: make([]func(res *http.Response), 0),
	}
}
