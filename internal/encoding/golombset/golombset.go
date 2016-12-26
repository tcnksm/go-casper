package golombset

import (
	"io"
	"math"

	"github.com/tcnksm/go-casper/internal/bits"
)

func Search(sc *Scanner, h uint) bool {
	var n uint
	for sc.Scan() {
		n += sc.Value()
		if h == n {
			return true
		}

		if h < n {
			return false
		}
	}

	return false
}

type Scanner struct {
	n uint // number of elements
	p uint // false-positive probability

	bitLen int // log2(P)

	value uint  // last scanned value
	err   error // first error while scanning

	r *bits.Reader
}

func NewScanner(r io.Reader, n, p uint) *Scanner {
	return &Scanner{
		n:      n,
		p:      p,
		bitLen: int(math.Log2(float64(p))),
		r:      bits.NewReader(r),
	}
}

// Scan advances the Scanner to the next value, which will then be
// available through the Value method. It will returns false when
// the scan stop, either by reaching the end of the input or an error.
func (sc *Scanner) Scan() bool {

	// Stop if last scan reached EOF
	if sc.err == io.EOF {
		return false
	}

	var v uint
	for {

		b, err := sc.r.Read(1)
		if err != nil {
			sc.err = err
			if err != io.EOF {
				return false
			}

		}

		if b == 0 {
			break
		}
		v += sc.p
	}

	r, err := sc.r.Read(sc.bitLen)
	if err != nil {
		sc.err = err
		if err != io.EOF {
			return false
		}
	}
	v += r

	sc.value = v
	return true
}

// Value returns last scanned value
func (sc *Scanner) Value() uint {
	return sc.value
}

// Err returns non-EOF error
func (sc *Scanner) Err() error {
	if sc.err == io.EOF {
		return nil
	}
	return sc.err
}

type Encoder struct {
	n uint // number of elements
	p uint // false-positive probability

	// bitLen is number of bits for writing remainder.
	bitLen int

	w io.Writer
}

// NewEncoder returns a new Encoder.
func NewEncoder(w io.Writer, n, p uint) *Encoder {
	bitLen := int(math.Log2(float64(p)))
	return &Encoder{
		n:      n,
		p:      p,
		bitLen: bitLen,
		w:      w,
	}
}

// Encode encodes the given array and write to underlying io.Writer.
// src must be the array of difference of uniformly distribute set of values.
func (e *Encoder) Encode(src []uint) error {
	if len(src) == 0 {
		return nil
	}

	wr := bits.NewWriter(e.w)
	for _, v := range src {
		q, r := v/e.p, v%e.p

		// Write unary code of quotient
		if err := wr.Write(1<<(uint(q)+1)-2, int(q+1)); err != nil {
			return err
		}

		// Write remainder
		if err := wr.Write(r, e.bitLen); err != nil {
			return err
		}
	}

	return wr.Flush()
}
