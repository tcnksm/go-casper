package casper

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/http2"
)

func TestPush(t *testing.T) {
	casper := New(1<<6, 10)
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opts := &Options{
			skipPush: true,
		}
		if err := casper.Push(w, r, "/static/example.jpg", opts); err != nil {
			t.Fatalf("Push failed: %s", err)
		}

		w.Header().Add("Content-Type", "text/html")
		w.Write([]byte(`<img src="/static/example.jpg"/>`))
		w.WriteHeader(http.StatusOK)
	}))

	if err := http2.ConfigureServer(ts.Config, nil); err != nil {
		t.Fatalf("Failed to configure h2 server: %s", err)
	}
	ts.TLS = ts.Config.TLSConfig
	ts.StartTLS()
	defer ts.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	if err := http2.ConfigureTransport(tr); err != nil {
		t.Fatalf("Failed to configure h2 transport: %s", err)
	}

	client := http.Client{
		Transport: tr,
	}

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if casper.alreadyPushed {
		t.Fatalf("content should not be already pushed")
	}

	cookies := res.Cookies()
	var exist bool
	for _, cookie := range cookies {
		if cookie.Name == cookieName {
			exist = true
			if got, want := cookie.Value, "%E5%00"; got != want {
				t.Fatalf("Get cookie %q, want %q", got, want)
			}
		}
	}

	if !exist {
		t.Fatalf("cookie %q is not set", cookieName)
	}
}

func TestGenerateCookie(t *testing.T) {
	// TODO(tcnksm): Check this value is reasonable or not
	want := "%16%D0%900O%3C%CE%B81%10"
	contents := []string{
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
	}

	casper := New(1<<6, len(contents))
	hashs := make([]uint, 0, len(contents))

	for _, content := range contents {
		hashs = append(hashs, casper.hash([]byte(content)))
	}

	got, err := casper.generateCookie(hashs)
	if err != nil {
		t.Fatalf("generateCookie should not fail")
	}

	if got != want {
		t.Fatalf("generateCookie=%q, want=%q", got, want)
	}
}

func TestPush_ServerPushNotSupported(t *testing.T) {
	var err error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cspr := New(1<<6, 10)
		err = cspr.Push(w, r, "/static/example.jpg", nil)
	}))
	defer ts.Close()

	http.Get(ts.URL)

	if err == nil {
		t.Fatal("expect to be failed") // TODO(tcnksm): define error
	}
}
