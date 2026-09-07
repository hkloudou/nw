// Package nw is a small HTTP client scaffold for API and scraping work:
// one Client, a JSON-serialisable cookie Jar, and plain (value, error) data flow.
package nw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// UserAgent is the default User-Agent header sent by New().
const UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

// Option mutates a request before it is sent. Client-level options (Use)
// run first, then per-call options, so per-call options win.
type Option func(*http.Request)

// Header sets a request header.
func Header(key, value string) Option {
	return func(r *http.Request) { r.Header.Set(key, value) }
}

// Query sets a URL query parameter.
func Query(key, value string) Option {
	return func(r *http.Request) {
		q := r.URL.Query()
		q.Set(key, value)
		r.URL.RawQuery = q.Encode()
	}
}

// Client wraps an *http.Client with a persistent Jar and default Options.
type Client struct {
	HTTP  *http.Client
	Jar   *Jar
	Debug bool // log every request and response body to the standard logger
	opts  []Option
}

// New returns a Client with an empty Jar, a 60s timeout and a browser User-Agent.
func New() *Client {
	jar := &Jar{}
	c := &Client{HTTP: &http.Client{Jar: jar, Timeout: 60 * time.Second}, Jar: jar}
	return c.Use(Header("User-Agent", UserAgent))
}

// Use adds Options applied to every request. Call it during setup, not concurrently with Do.
func (c *Client) Use(opts ...Option) *Client {
	c.opts = append(c.opts, opts...)
	return c
}

// Proxy routes all traffic through proxyURL. An empty string restores the
// default behaviour (HTTP_PROXY / HTTPS_PROXY from the environment).
func (c *Client) Proxy(proxyURL string) error {
	t := http.DefaultTransport.(*http.Transport).Clone()
	if proxyURL != "" {
		u, err := url.Parse(proxyURL)
		if err != nil {
			return err
		}
		t.Proxy = http.ProxyURL(u)
	}
	c.HTTP.Transport = t
	return nil
}

// Get sends a GET request.
func (c *Client) Get(ctx context.Context, rawURL string, opts ...Option) (*Response, error) {
	return c.Send(ctx, http.MethodGet, rawURL, nil, opts...)
}

// PostJSON sends body as application/json. A string or []byte body is sent
// as-is; anything else goes through json.Marshal.
func (c *Client) PostJSON(ctx context.Context, rawURL string, body any, opts ...Option) (*Response, error) {
	var b []byte
	switch v := body.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		var err error
		if b, err = json.Marshal(v); err != nil {
			return nil, err
		}
	}
	opts = append([]Option{Header("Content-Type", "application/json")}, opts...)
	return c.Send(ctx, http.MethodPost, rawURL, bytes.NewReader(b), opts...)
}

// PostForm sends form as application/x-www-form-urlencoded.
func (c *Client) PostForm(ctx context.Context, rawURL string, form url.Values, opts ...Option) (*Response, error) {
	opts = append([]Option{Header("Content-Type", "application/x-www-form-urlencoded")}, opts...)
	return c.Send(ctx, http.MethodPost, rawURL, strings.NewReader(form.Encode()), opts...)
}

// Send builds a request with any method and body and passes it to Do.
func (c *Client) Send(ctx context.Context, method, rawURL string, body io.Reader, opts ...Option) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req, opts...)
}

// Stream applies the Options and sends req, returning the raw response
// with its body unread. The caller must close resp.Body. Use it for SSE,
// large downloads or anything that should not be buffered.
func (c *Client) Stream(req *http.Request, opts ...Option) (*http.Response, error) {
	for _, o := range c.opts {
		o(req)
	}
	for _, o := range opts {
		o(req)
	}
	return c.HTTP.Do(req)
}

// Do is Stream followed by reading the whole body.
//
// A transport failure returns (nil, err). A status >= 400 returns the
// Response together with a *StatusError, so the body is still inspectable.
func (c *Client) Do(req *http.Request, opts ...Option) (*Response, error) {
	start := time.Now()
	r, err := c.Stream(req, opts...)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if c.Debug {
		log.Printf("nw: %s %s -> %d (%s)\n%s", req.Method, req.URL, r.StatusCode, time.Since(start).Round(time.Millisecond), body)
	}
	resp := &Response{Response: r, Body: body}
	if r.StatusCode >= 400 {
		return resp, &StatusError{resp}
	}
	return resp, nil
}

// Response is a fully-read HTTP response.
type Response struct {
	*http.Response
	Body []byte
}

// String returns the body as text.
func (r *Response) String() string { return string(r.Body) }

// JSON decodes the body into v.
func (r *Response) JSON(v any) error { return json.Unmarshal(r.Body, v) }

// StatusError is returned by Do for responses with status >= 400.
type StatusError struct{ *Response }

func (e *StatusError) Error() string {
	return fmt.Sprintf("nw: %s %s: %s", e.Request.Method, e.Request.URL, e.Status)
}

// JSON decodes a (Response, error) pair straight into a T:
//
//	user, err := nw.JSON[User](c.Get(ctx, url))
func JSON[T any](r *Response, err error) (T, error) {
	var v T
	if err != nil {
		return v, err
	}
	return v, r.JSON(&v)
}
