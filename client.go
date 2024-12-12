package nw

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/hkloudou/nw/v2/cookiejar"
)

type Client struct {
	_localStorage map[string]string
	httpClient    *http.Client
	cookieJar     *cookiejar.Jar
	userAgent     string
}

func NewClient() *Client {
	return &Client{
		_localStorage: map[string]string{},
		cookieJar:     cookiejar.NewJar(&cookiejar.Options{}),
		userAgent:     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	}
}

func NewFromData(localStorage map[string]string, jar *cookiejar.Jar) *Client {
	return &Client{
		_localStorage: localStorage,
		cookieJar:     jar,
		userAgent:     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	}
}

func (m *Client) Storages() map[string]string {
	return m._localStorage
}

func (m *Client) LocalStorageGet(key string) string {
	result, found := m._localStorage[key]
	if !found {
		return ""
	}
	return result
}

func (m *Client) LocalStorageSet(key, value string) {
	m._localStorage[key] = value
}

// GetHTTPClient 获得GetHTTPClient对象
func (m *Client) GetHTTPClient() *http.Client {
	if m.httpClient == nil {
		m.httpClient = &http.Client{
			Jar: m.cookieJar,
		}
	}
	return m.httpClient
}

func (m *Client) String() string {
	b, _ := json.Marshal(m.cookieJar)
	return string(b)
}

func (me *Client) WiseProxy(proxyURL *url.URL) {
	client := me.GetHTTPClient()
	if proxyURL != nil {
		client.Transport = &http.Transport{
			// DisableKeepAlives: true,
			Proxy: func(req *http.Request) (*url.URL, error) {
				return proxyURL, nil
			},
			IdleConnTimeout: 2 * time.Minute,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	} else {
		client.Transport = &http.Transport{
			// DisableKeepAlives: true,
			Proxy:           http.ProxyFromEnvironment,
			IdleConnTimeout: 2 * time.Minute,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
}

func (me *Client) SetTimeOut(timeout time.Duration) {
	client := me.GetHTTPClient()
	client.Timeout = timeout
}
func (me *Client) Close() {
	client := me.GetHTTPClient()
	client.CloseIdleConnections()
}

func (me *Client) GetJar() *cookiejar.Jar {
	return me.cookieJar
}

func (me *Client) SetJar(tmp *cookiejar.Jar) {
	me.cookieJar.DeepCopyFrom(tmp)
}

func (me *Client) SetUserAgent(str string) string {
	me.userAgent = str
	return me.userAgent
}
