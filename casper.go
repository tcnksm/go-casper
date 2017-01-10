package casper

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/tcnksm/go-casper/internal/encoding/golomb"
)

const (
	defaultCookieName = "x-go-casper"
)

// Casper stores
type Casper struct {
	p uint
	n uint

	// alreadyPushed is true when the request cookie indicates
	// the given content is already pushed.
	alreadyPushed bool
}

// Options includes casper push options.
type Options struct {
	*http.PushOptions

	// skipPush skips server pushing. This should be only used in testing.
	// Currently, it's kinda hard to receive http push in go http client.
	// Set this and ignore pushing and only test cookie part.
	// This should be removed in future.
	skipPush bool
}

// New initialize casper.
func New(p, n int) *Casper {
	return &Casper{
		p: uint(p),
		n: uint(n),
	}
}

// Push pushes the given content and set cookie value.
func (c *Casper) Push(w http.ResponseWriter, r *http.Request, content string, opts *Options) error {
	// Pusher is used later in this function but should check
	// it's available or not first to avoid waste calc.
	pusher, ok := w.(http.Pusher)
	if !ok {
		// go1.8 or later
		return errors.New("server push is not supported")
	}

	hashs, err := c.decodeCookie(r)
	if err != nil {
		return err
	}

	h := c.hash([]byte(content))
	if search(hashs, h) {
		c.alreadyPushed = true
		return nil
	}

	if !opts.skipPush {
		if err := pusher.Push(content, opts.PushOptions); err != nil {
			return err
		}
	}

	hashs = append(hashs, h)
	cookieValue, err := c.generateCookie(hashs)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:  defaultCookieName,
		Value: cookieValue,
	}
	http.SetCookie(w, cookie)

	return nil
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

// generateCookie generates cookie value from the given hash array.
func (c *Casper) generateCookie(hashs []uint) (string, error) {

	sort.Slice(hashs, func(i, j int) bool {
		return hashs[i] < hashs[j]
	})

	var buf bytes.Buffer
	if err := golomb.Encode(&buf, hashs, c.p); err != nil {
		return "", err
	}

	return url.QueryEscape(buf.String()), nil
}

// decodeCookie reads cookie from http request and decode it to hash array.
func (c *Casper) decodeCookie(r *http.Request) ([]uint, error) {
	cookie, err := r.Cookie(defaultCookieName)
	if err != nil && err != http.ErrNoCookie {
		return nil, err
	}

	if err == http.ErrNoCookie {
		// TODO(tcnksm): Set size
		var hashs []uint
		return hashs, nil
	}

	value, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return nil, err
	}

	hashs, err := golomb.DecodeAll([]byte(value), c.p)
	if err != nil {
		return nil, err
	}

	return hashs, nil
}

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
