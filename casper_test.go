package casper

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/http2"
)

func NewPushServer(t *testing.T, casper *Casper, content string) *httptest.Server {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: "19881124",
			Path:  "/",
		})

		opts := &Options{
			skipPush: true,
		}
		if _, err := casper.Push(w, r, content, opts); err != nil {
			t.Fatalf("Push failed: %s", err)
		}

		w.Header().Add("Content-Type", "text/html")
		w.Write([]byte(""))
		w.WriteHeader(http.StatusOK)
	}))

	if err := http2.ConfigureServer(ts.Config, nil); err != nil {
		t.Fatalf("Failed to configure h2 server: %s", err)
	}
	ts.TLS = ts.Config.TLSConfig
	ts.StartTLS()

	return ts
}

func h2Client(t *testing.T) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	if err := http2.ConfigureTransport(tr); err != nil {
		t.Fatalf("Failed to configure h2 transport: %s", err)
	}

	return &http.Client{
		Transport: tr,
	}
}

func TestPush(t *testing.T) {
	casper := New(1<<6, 10)
	content := "/static/example.jpg"

	ts := NewPushServer(t, casper, content)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	client := h2Client(t)
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if casper.alreadyPushed {
		t.Fatalf("content should not be already pushed")
	}

	cookies := res.Cookies()
	if got, want := len(cookies), 2; got != want {
		t.Fatalf("Number of cookie %d, want %d", got, want)
	}

	if got, want := cookies[0].Name, "session"; got != want {
		t.Fatalf("Get cookie name %q, want %q", got, want)
	}

	if got, want := cookies[0].Value, "19881124"; got != want {
		t.Fatalf("Get cookie value %q, want %q", got, want)
	}

	if got, want := cookies[1].Name, cookieName; got != want {
		t.Fatalf("Get cookie name %q, want %q", got, want)
	}

	if got, want := cookies[1].Value, "5QA="; got != want {
		t.Fatalf("Get cookie value %q, want %q", got, want)
	}
}

func TestPushWithCookie(t *testing.T) {
	casper := New(1<<6, 10)
	content := "/static/example.jpg"

	ts := NewPushServer(t, casper, content)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	cookie := &http.Cookie{
		Name:  cookieName,
		Value: "5QA=",
	}
	req.AddCookie(cookie)

	client := h2Client(t)
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if !casper.alreadyPushed {
		t.Fatalf("%q should be cached", content)
	}

	cookies := res.Cookies()
	if got, want := len(cookies), 2; got != want {
		t.Fatalf("Number of cookie %d, want %d", got, want)
	}

	if got, want := len(cookies), 2; got != want {
		t.Fatalf("Number of cookie %d, want %d", got, want)
	}

	if got, want := cookies[0].Name, "session"; got != want {
		t.Fatalf("Get cookie name %q, want %q", got, want)
	}

	if got, want := cookies[0].Value, "19881124"; got != want {
		t.Fatalf("Get cookie value %q, want %q", got, want)
	}

	if got, want := cookies[1].Name, cookieName; got != want {
		t.Fatalf("Get cookie name %q, want %q", got, want)
	}

	if got, want := cookies[1].Value, "5QA="; got != want {
		t.Fatalf("Get cookie value %q, want %q", got, want)
	}
}

func TestPushMultipleContents(t *testing.T) {
	casper := New(1<<6, 2)
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contents := []string{
			"/js/jquery-1.9.1.min.js",
			"/assets/style.css",
		}
		opts := &Options{
			skipPush: true,
		}

		r, err := casper.Push(w, r, contents[0], opts)
		if err != nil {
			t.Fatalf("Push failed: %s", err)
		}

		r, err = casper.Push(w, r, contents[1], opts)
		if err != nil {
			t.Fatalf("Push failed: %s", err)
		}

		w.Header().Add("Content-Type", "text/html")
		w.Write([]byte(""))
		w.WriteHeader(http.StatusOK)
	}))

	if err := http2.ConfigureServer(ts.Config, nil); err != nil {
		t.Fatalf("Failed to configure h2 server: %s", err)
	}
	ts.TLS = ts.Config.TLSConfig
	ts.StartTLS()

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	client := h2Client(t)
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if casper.alreadyPushed {
		t.Fatalf("content should not be already pushed")
	}

	cookies := res.Cookies()
	if got, want := len(cookies), 1; got != want {
		t.Fatalf("Number of cookie %d, want %d", got, want)
	}

	if got, want := cookies[0].Name, cookieName; got != want {
		t.Fatalf("Get cookie name %q, want %q", got, want)
	}

	if got, want := cookies[0].Value, "gU4="; got != want {
		t.Fatalf("Get cookie value %q, want %q", got, want)
	}

}

func TestGenerateCookie(t *testing.T) {

	cases := []struct {
		contents    []string
		P           int
		cookieValue string
	}{
		{
			[]string{
				"/js/jquery-1.9.1.min.js",
				"/assets/style.css",
			},
			1 << 6,
			"gU4=",
		},
		{
			[]string{
				"/static/example1.jpg",
				"/static/example2.jpg",
				"/static/example3.jpg",
				"/static/example4.jpg",
				"/static/example5.jpg",
				"/static/example6.jpg",
				"/static/example7.jpg",
				"/static/example8.jpg",
				"/static/example9.jpg",
				"/static/example10.jpg",
			},
			1 << 6,
			// Minimum number of bits is N*log(P)
			// = 10 * log(1<<6) = 60 bits = 7.5bytes
			"FtCQME88zrgxEA==", // 16 bytes
		},
	}

	for _, tc := range cases {
		casper := New(tc.P, len(tc.contents))

		hashValues := make([]uint, 0, len(tc.contents))
		for _, content := range tc.contents {
			hashValues = append(hashValues, casper.hash([]byte(content)))
		}

		cookie, err := casper.generateCookie(hashValues)
		if err != nil {
			t.Fatalf("generateCookie should not fail")
		}

		if got, want := cookie.Value, tc.cookieValue; got != want {
			t.Fatalf("generateCookie=%q, want=%q", got, want)
		}
	}
}

func TestPush_ServerPushNotSupported(t *testing.T) {
	var err error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cspr := New(1<<6, 10)
		_, err = cspr.Push(w, r, "/static/example.jpg", nil)
	}))
	defer ts.Close()

	http.Get(ts.URL)

	if err == nil {
		t.Fatal("expect to be failed") // TODO(tcnksm): define error
	}
}
