package casper

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/tcnksm/go-casper/internal/encoding/golombset"
)

const (
	cookieName = "x-go-casper"
)

type Casper struct {
	p uint
	n uint // TODO(tcnksm): need ..?
}

type Options struct {

	// skipPush skips server pushing. This should be only used in testing.
	// Currently, it's kinda hard to receive http push in go http client.
	// Set this and ignore pushing and only test cookie part.
	// This should be removed in future.
	skipPush bool
}

func New(p, n uint) *Casper {
	return &Casper{
		p: p,
		n: n,
	}
}

// Push pushes
//
// TODO(tcnksm): Handle multiple contents
func (c *Casper) Push(w http.ResponseWriter, r *http.Request, content string, opts *Options) error {
	pusher, ok := w.(http.Pusher)
	if !ok {
		// go1.8 or later
		return errors.New("server push is not supported")
	}

	hashs, err := c.readCookie(r)
	if err != nil {
		return err
	}

	// TODO(tcnksm): Enable to configure?
	hashF := func(v []byte) uint {
		h := md5.New()
		h.Write(v)
		b := h.Sum(nil)

		s := hex.EncodeToString(b[12:16])
		i, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			panic(err)
		}
		return uint(i) % (c.n * c.p)
	}

	// TODO(tcnksm): binary search (or enable to configure?)
	searchF := func(a []uint, h uint) bool {
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

	h := hashF([]byte(content))
	if searchF(hashs, h) {
		log.Println("Already pushded")
		return nil
	}

	if !opts.skipPush {
		if err := pusher.Push(content, nil); err != nil {
			return err
		}
		log.Println("Pushed")
	}

	// Set cookie
	hashs = append(hashs, h)
	sort.Slice(hashs, func(i, j int) bool {
		return hashs[i] < hashs[j]
	})

	var buf bytes.Buffer
	if err := golombset.Encode(&buf, hashs, c.p); err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:  cookieName,
		Value: url.QueryEscape(buf.String()),
	}
	http.SetCookie(w, cookie)

	return nil
}

// readCookie reads cookie from http request and decode it to hash array.
func (c *Casper) readCookie(r *http.Request) ([]uint, error) {
	cookie, err := r.Cookie(cookieName)
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

	hashs, err := golombset.DecodeAll([]byte(value), c.p)
	if err != nil {
		return nil, err
	}

	return hashs, nil
}
