package casper

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/tcnksm/go-casper/internal/encoding/golomb"
)

const (
	cookieName = "x-go-casper"
)

var (
	hashContextkey = &contextKey{"casper-hash"}
)

// Casper stores
type Casper struct {
	p uint
	n uint

	// inMemory decides executing actual server push or fakely push
	// to in memory buffer(buf) for testing. This should be true only in testing.
	//
	// Currently, it's kinda hard to receive http push in go http client.
	// This should be removed in future.
	inMemory bool
	buf      []string
}

// Options includes casper push options.
type Options struct {
	*http.PushOptions
}

type contextKey struct {
	name string
}

func (c *contextKey) String() string {
	return c.name
}

// New initialize casper.
func New(p, n int) *Casper {
	return &Casper{
		p: uint(p),
		n: uint(n),
	}
}

// Push pushes the given content and set cookie value.
func (c *Casper) Push(w http.ResponseWriter, r *http.Request, contents []string, opts *Options) (*http.Request, error) {
	// Pusher is used later in this function but should check
	// it's available or not first to avoid unnessary calc.
	pusher, ok := w.(http.Pusher)
	if !ok {
		return r, errors.New("server push is not supported") // go1.8 or later
	}

	// Remove casper cookie header if it's already exists.
	if cookies, ok := w.Header()["Set-Cookie"]; ok && len(cookies) != 0 {
		w.Header().Del("Set-Cookie")
		for _, cookieStr := range cookies {
			if strings.Contains(cookieStr, cookieName+"=") {
				continue
			}
			w.Header().Add("Set-Cookie", cookieStr)
		}
	}

	// Get hash values assosiated with previous parent context.
	// If none, then read it from the request cookie.
	hashValues := contextHashValues(r.Context())
	if hashValues == nil {
		var err error
		hashValues, err = c.readCookie(r)
		if err != nil {
			return r, err
		}
	}

	// Push contents one by one.
	for _, content := range contents {
		h := c.hash([]byte(content))
		if search(hashValues, h) {
			continue
		}

		if c.inMemory {
			// Push to in memory buffer. This is only for testing.
			c.buf = append(c.buf, content)
		} else {
			if err := pusher.Push(content, opts.PushOptions); err != nil {
				return r, err
			}
		}

		hashValues = append(hashValues, h)
	}

	// TODO(tcnksm): Can be skip when nothing is pushed.
	cookie, err := c.generateCookie(hashValues)
	if err != nil {
		return r, err
	}
	http.SetCookie(w, cookie)

	return r.WithContext(withHashValues(r.Context(), hashValues)), nil
}

// hash generate a hash value from the given bytes for
// n elements and p faslse positive probability.
//
// It's ok to use md5 since we just need a hash that generates
// uniformally-distributed values for best results.
func (c *Casper) hash(p []byte) uint {
	h := md5.New()
	h.Write(p)
	b := h.Sum(nil)

	s := hex.EncodeToString(b[12:16])
	i, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		panic(err)
	}
	return uint(i) % (c.n * c.p)
}

// generateCookie generates cookie from the given hash values.
func (c *Casper) generateCookie(hashValues []uint) (*http.Cookie, error) {

	// golomb encoder expect the given array is sorted.
	sort.Slice(hashValues, func(i, j int) bool {
		return hashValues[i] < hashValues[j]
	})

	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.URLEncoding, &buf)
	if err := golomb.Encode(encoder, hashValues, c.p); err != nil {
		return nil, err
	}

	if err := encoder.Close(); err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:  cookieName,
		Value: buf.String(),
	}, nil
}

// readCookie reads cookie from http request and decode it to hash array.
func (c *Casper) readCookie(r *http.Request) ([]uint, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil && err != http.ErrNoCookie {
		return nil, err
	}

	if err == http.ErrNoCookie {
		hashValues := make([]uint, 0, c.n)
		return hashValues, nil
	}

	// Decode golomb coded cookie value to original hash values array.
	decoder := base64.NewDecoder(base64.URLEncoding, strings.NewReader(cookie.Value))
	hashValues, err := golomb.DecodeAll(decoder, c.p)
	if err != nil {
		return nil, err
	}

	return hashValues, nil
}

// withHashValues returns a new context based on previsous parent context.
// It sets hashValues which is used for generating golomb encoded cookie value.
func withHashValues(parent context.Context, hashValues []uint) context.Context {
	return context.WithValue(parent, hashContextkey, hashValues)
}

// contextHashValues returns the hashValues assosiated with the
// provided context. If none, it returns nil,
func contextHashValues(ctx context.Context) []uint {
	hashValues, _ := ctx.Value(hashContextkey).([]uint)
	return hashValues
}

// search looks up the provided slices contains the given value.
//
// TODO(tcnksm): binary search (or enable to configure?)
func search(a []uint, h uint) bool {
	for i := 0; i < len(a); i++ {
		if h == a[i] {
			return true
		}

		if h < a[i] {
			return false
		}
	}
	return false
}
