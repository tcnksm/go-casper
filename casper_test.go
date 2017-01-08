package casper

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/http2"
)

func TestPush(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cspr := New(1<<6, 10)
		opts := &Options{
			skipPush: true,
		}
		if err := cspr.Push(w, r, "/static/example.jpg", opts); err != nil {
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
