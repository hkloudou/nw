package nw

import (
	"cmp"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

// Jar is an http.CookieJar whose contents serialise to JSON, so a logged-in
// session can be saved and reused across processes. The zero value is ready to use.
//
// Cookies are stored with the Netscape convention: a Domain starting with "."
// is shared with subdomains, otherwise the cookie is host-only. No public
// suffix list is consulted; a server may set cookies for itself or any parent
// domain, which is what API and scraping clients want.
type Jar struct {
	mu      sync.Mutex
	cookies map[string]*http.Cookie // keyed by domain;path;name
}

// SetCookies implements http.CookieJar.
func (j *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	host, now := hostname(u), time.Now()
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.cookies == nil {
		j.cookies = map[string]*http.Cookie{}
	}
	for _, in := range cookies {
		c := *in
		c.Raw, c.RawExpires, c.Unparsed = "", "", nil
		if c.Domain == "" {
			c.Domain = host
		} else {
			d := strings.TrimPrefix(strings.ToLower(c.Domain), ".")
			if d != host && !strings.HasSuffix(host, "."+d) {
				continue // only the host itself or a parent domain is allowed
			}
			c.Domain = "." + d
		}
		if !strings.HasPrefix(c.Path, "/") {
			c.Path = defaultPath(u.Path)
		}
		if c.MaxAge > 0 {
			c.Expires, c.MaxAge = now.Add(time.Duration(c.MaxAge)*time.Second), 0
		}
		if c.MaxAge < 0 || expired(&c, now) {
			delete(j.cookies, key(&c))
		} else {
			j.cookies[key(&c)] = &c
		}
	}
}

// Cookies implements http.CookieJar.
func (j *Jar) Cookies(u *url.URL) []*http.Cookie {
	host, path, now := hostname(u), u.Path, time.Now()
	if path == "" {
		path = "/"
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	var out []*http.Cookie
	for k, c := range j.cookies {
		if expired(c, now) {
			delete(j.cookies, k)
			continue
		}
		if c.Secure && u.Scheme != "https" || !domainMatch(c.Domain, host) || !pathMatch(c.Path, path) {
			continue
		}
		out = append(out, c)
	}
	// RFC 6265 §5.4: longer paths first.
	slices.SortFunc(out, func(a, b *http.Cookie) int {
		return cmp.Or(len(b.Path)-len(a.Path), strings.Compare(a.Name, b.Name))
	})
	for i, c := range out {
		out[i] = &http.Cookie{Name: c.Name, Value: c.Value, Quoted: c.Quoted}
	}
	return out
}

// All returns a sorted copy of every stored cookie with full attributes.
func (j *Jar) All() []*http.Cookie {
	j.mu.Lock()
	defer j.mu.Unlock()
	out := make([]*http.Cookie, 0, len(j.cookies))
	for _, c := range j.cookies {
		cc := *c
		out = append(out, &cc)
	}
	slices.SortFunc(out, func(a, b *http.Cookie) int { return strings.Compare(key(a), key(b)) })
	return out
}

// Clear removes every cookie.
func (j *Jar) Clear() {
	j.mu.Lock()
	j.cookies = nil
	j.mu.Unlock()
}

// MarshalJSON encodes the jar as a JSON array of cookies.
func (j *Jar) MarshalJSON() ([]byte, error) { return json.Marshal(j.All()) }

// UnmarshalJSON merges a JSON array of cookies (as written by MarshalJSON) into the jar.
func (j *Jar) UnmarshalJSON(b []byte) error {
	var cs []*http.Cookie
	if err := json.Unmarshal(b, &cs); err != nil {
		return err
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.cookies == nil {
		j.cookies = map[string]*http.Cookie{}
	}
	for _, c := range cs {
		j.cookies[key(c)] = c
	}
	return nil
}

// Save writes the jar to a JSON file.
func (j *Jar) Save(path string) error {
	b, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

// Load merges a JSON file written by Save into the jar.
func (j *Jar) Load(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, j)
}

// key identifies a cookie by domain, path and name (RFC 6265 §5.3 step 11); a
// host-only cookie and a domain cookie for the same host replace each other.
func key(c *http.Cookie) string { return strings.TrimPrefix(c.Domain, ".") + ";" + c.Path + ";" + c.Name }

func hostname(u *url.URL) string { return strings.ToLower(strings.TrimSuffix(u.Hostname(), ".")) }

func expired(c *http.Cookie, now time.Time) bool { return !c.Expires.IsZero() && !c.Expires.After(now) }

func domainMatch(domain, host string) bool {
	if !strings.HasPrefix(domain, ".") {
		return domain == host
	}
	return host == domain[1:] || strings.HasSuffix(host, domain)
}

// pathMatch is RFC 6265 §5.1.4 path-match.
func pathMatch(cookiePath, path string) bool {
	return path == cookiePath || strings.HasPrefix(path, cookiePath) &&
		(strings.HasSuffix(cookiePath, "/") || path[len(cookiePath)] == '/')
}

// defaultPath is RFC 6265 §5.1.4 default-path: the request path up to, but not including, its last "/".
func defaultPath(path string) string {
	if i := strings.LastIndex(path, "/"); i > 0 {
		return path[:i]
	}
	return "/"
}
