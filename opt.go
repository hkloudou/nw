package nw

import (
	"io"
	"net/http"
	"time"
)

type nwOption struct {
	dataKeys   []string
	codeKeys   []string
	msgKeys    []string
	site       string
	client     *http.Client
	postReader io.Reader
	log        bool
}

type NwOption func(*nwOption)

func WithClient(client *http.Client) NwOption {
	return func(o *nwOption) {
		o.client = client
	}
}

func WithLog(uselog bool) NwOption {
	return func(o *nwOption) {
		o.log = uselog
	}
}

func WithPostData(reader io.Reader) NwOption {
	return func(o *nwOption) {
		o.postReader = reader
	}
}

func WithSite(site string) NwOption {
	return func(o *nwOption) {
		o.site = site
	}
}

func WithCodeKeys(keys ...string) NwOption {
	return func(o *nwOption) {
		o.codeKeys = keys
	}
}

func WithMsgKeys(keys ...string) NwOption {
	return func(o *nwOption) {
		o.msgKeys = keys
	}
}

func WithDataKeys(keys ...string) NwOption {
	return func(o *nwOption) {
		o.dataKeys = keys
	}
}

func getOption(opts ...NwOption) *nwOption {
	var o = &nwOption{
		codeKeys: []string{"c", "Code", "errcode"},
		msgKeys:  []string{"m", "Msg", "Message", "errmsg"},
		dataKeys: []string{"d", "Data", "data", ""},
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.client == nil {
		o.client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}
	return o
}
