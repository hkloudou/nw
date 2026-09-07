package nw

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"
)

func TestClientJSONAndOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"method": r.Method,
			"ua":     r.UserAgent(),
			"token":  r.Header.Get("X-Token"),
			"page":   r.URL.Query().Get("page"),
			"ctype":  r.Header.Get("Content-Type"),
		})
	}))
	defer srv.Close()

	type echo struct{ Method, UA, Token, Page, Ctype string }
	c := New().Use(Header("X-Token", "client"))
	ctx := context.Background()

	got, err := JSON[echo](c.Get(ctx, srv.URL, Query("page", "2")))
	if err != nil {
		t.Fatal(err)
	}
	if got.Method != "GET" || got.UA != UserAgent || got.Token != "client" || got.Page != "2" {
		t.Fatalf("unexpected echo: %+v", got)
	}

	// Per-call options override client-level ones.
	got, err = JSON[echo](c.PostJSON(ctx, srv.URL, map[string]int{"a": 1}, Header("X-Token", "call")))
	if err != nil || got.Method != "POST" || got.Token != "call" || got.Ctype != "application/json" {
		t.Fatalf("PostJSON: %+v %v", got, err)
	}

	got, err = JSON[echo](c.PostForm(ctx, srv.URL, url.Values{"a": {"1"}}))
	if err != nil || got.Ctype != "application/x-www-form-urlencoded" {
		t.Fatalf("PostForm: %+v %v", got, err)
	}
}

func TestStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusTeapot)
	}))
	defer srv.Close()

	resp, err := New().Get(context.Background(), srv.URL)
	var se *StatusError
	if !errors.As(err, &se) || se.StatusCode != http.StatusTeapot {
		t.Fatalf("want StatusError 418, got %v", err)
	}
	if resp == nil || resp.String() != "nope\n" {
		t.Fatalf("body should still be readable, got %v", resp)
	}
	if _, err := JSON[map[string]any](resp, err); !errors.Is(err, se) {
		t.Fatalf("JSON must pass the error through, got %v", err)
	}
}

func TestCookieRoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "sid", Value: "s3cret", Path: "/", MaxAge: 3600})
		default:
			if c, err := r.Cookie("sid"); err == nil {
				w.Write([]byte(c.Value))
			}
		}
	}))
	defer srv.Close()
	ctx := context.Background()

	c := New()
	if _, err := c.Get(ctx, srv.URL+"/login"); err != nil {
		t.Fatal(err)
	}
	if resp, _ := c.Get(ctx, srv.URL+"/me"); resp.String() != "s3cret" {
		t.Fatalf("cookie not sent back: %q", resp.String())
	}

	// Persist to disk and restore into a brand-new client.
	path := filepath.Join(t.TempDir(), "cookies.json")
	if err := c.Jar.Save(path); err != nil {
		t.Fatal(err)
	}
	c2 := New()
	if err := c2.Jar.Load(path); err != nil {
		t.Fatal(err)
	}
	if resp, _ := c2.Get(ctx, srv.URL+"/me"); resp.String() != "s3cret" {
		t.Fatalf("restored cookie not sent back: %q", resp.String())
	}
	all := c2.Jar.All()
	if len(all) != 1 || all[0].Name != "sid" || all[0].Expires.IsZero() || all[0].MaxAge != 0 {
		t.Fatalf("MaxAge should be stored as absolute Expires: %+v", all)
	}

	c2.Jar.Clear()
	if resp, _ := c2.Get(ctx, srv.URL+"/me"); resp.String() != "" {
		t.Fatalf("Clear did not drop cookies: %q", resp.String())
	}
}

func TestJarMatching(t *testing.T) {
	mustURL := func(s string) *url.URL {
		u, err := url.Parse(s)
		if err != nil {
			t.Fatal(err)
		}
		return u
	}
	names := func(u string) (out []string) {
		for _, c := range jar.Cookies(mustURL(u)) {
			out = append(out, c.Name)
		}
		return out
	}
	eq := func(u string, want ...string) {
		t.Helper()
		got := names(u)
		if len(got) != len(want) {
			t.Fatalf("%s: got %v want %v", u, got, want)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("%s: got %v want %v", u, got, want)
			}
		}
	}

	jar.SetCookies(mustURL("https://www.example.com/a/b"), []*http.Cookie{
		{Name: "host", Value: "1"},                                   // host-only, default path /a
		{Name: "dom", Value: "1", Domain: ".example.com", Path: "/"}, // shared with subdomains
		{Name: "sec", Value: "1", Path: "/", Secure: true},
		{Name: "evil", Value: "1", Domain: "other.com"},   // rejected
		{Name: "gone", Value: "1", Path: "/", MaxAge: -1}, // deleted
		{Name: "old", Value: "1", Path: "/", Expires: time.Now().Add(-time.Hour)},
	})

	eq("https://www.example.com/a/b", "host", "dom", "sec") // path /a sorts before /
	eq("https://www.example.com/a", "host", "dom", "sec")
	eq("https://www.example.com/ab", "dom", "sec") // "/a" must not prefix-match "/ab"
	eq("http://www.example.com/a", "host", "dom")  // Secure only over https
	eq("https://api.example.com/", "dom")          // host-only cookie stays on www
	eq("https://example.com/", "dom")
	eq("https://other.com/")

	// Expiry is honoured lazily on read.
	jar.SetCookies(mustURL("https://example.com/"), []*http.Cookie{{Name: "ttl", Path: "/", MaxAge: 1}})
	eq("https://example.com/", "dom", "ttl")
	for _, c := range jar.All() {
		if c.Name == "ttl" && time.Until(c.Expires) > time.Second {
			t.Fatalf("MaxAge not converted to Expires: %+v", c)
		}
	}
	jar.mu.Lock()
	jar.cookies[key(&http.Cookie{Domain: "example.com", Path: "/", Name: "ttl"})].Expires = time.Now().Add(-time.Second)
	jar.mu.Unlock()
	eq("https://example.com/", "dom")
}

var jar Jar // exercise the zero value

func TestJarJSON(t *testing.T) {
	var j Jar
	if err := json.Unmarshal([]byte(`[{"Name":"a","Value":"1","Domain":".x.com","Path":"/"}]`), &j); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(&j)
	if err != nil {
		t.Fatal(err)
	}
	var back []http.Cookie
	if err := json.Unmarshal(b, &back); err != nil || len(back) != 1 || back[0].Domain != ".x.com" {
		t.Fatalf("round trip failed: %s %v", b, err)
	}
}

func TestJarReplaceAndQuoted(t *testing.T) {
	var j Jar
	u, _ := url.Parse("https://example.com/")
	j.SetCookies(u, []*http.Cookie{{Name: "sid", Value: "a", Path: "/"}})
	j.SetCookies(u, []*http.Cookie{{Name: "sid", Value: "b c", Path: "/", Domain: "example.com", Quoted: true}})
	got := j.Cookies(u)
	if len(got) != 1 || got[0].Value != "b c" || !got[0].Quoted {
		t.Fatalf("host-only and domain cookie with the same name must replace each other and keep Quoted: %+v", got)
	}
}

func TestStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Echo", r.Header.Get("X-Token"))
		for i := 0; i < 3; i++ {
			w.Write([]byte("chunk\n"))
			w.(http.Flusher).Flush()
		}
	}))
	defer srv.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	resp, err := New().Use(Header("X-Token", "t")).Stream(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.Header.Get("X-Echo") != "t" {
		t.Fatal("client options not applied")
	}
	n := 0
	for sc := bufio.NewScanner(resp.Body); sc.Scan(); n++ {
	}
	if n != 3 {
		t.Fatalf("want 3 lines, got %d", n)
	}
}
