package golomb

import (
	"bytes"
	"errors"
	"io"
	"math"

	"github.com/tcnksm/go-casper/internal/bits"
)

var errPadding = errors.New("padding")

// DecodeAll decodes...
func DecodeAll(src []byte, p uint) ([]uint, error) {
	// TODO(tcnksm): Receive dst from outside as argument.
	var dst []uint
	br := bits.NewReader(bytes.NewReader(src))
	prev := uint(0)
	for {
		v, err := decode(br, p)
		if err == errPadding {
			// Ignore padding value
			return dst, nil
		}

		if err == io.EOF {
			dst = append(dst, v+prev)
			return dst, nil
		}

		if err != nil {
			return nil, err
		}

		dst = append(dst, v+prev)

		prev = v + prev
	}
}

func decode(br *bits.Reader, p uint) (uint, error) {
	var v uint

	// Decode unary parts. Sum p until enconter 0 bits.
	for {
		b, err := br.Read(1)
		if err != nil {
			return 0, err
		}

		if b == 0 {
			break
		}
		v += p
	}

	// Decode remainder parts.
	bitLen := int(math.Log2(float64(p)))
	r, err := br.Read(bitLen)
	if err == io.EOF && r == 0 {
		return 0, errPadding
	}

	if err != nil && err != io.EOF {
		return 0, err
	}
	v += r

	return v, err
}

// Encode encodes the given uint array and writes to underlying writer.
// p is false-positive probability. The src array must be uniformly
// distribute set of values.
func Encode(w io.Writer, src []uint, p uint) error {
	if len(src) == 0 {
		return nil
	}

	// bitLen is number of bits for writing remainder.
	bitLen := int(math.Log2(float64(p)))

	// TODO(tcnksm)
	wr := bits.NewWriter(w)

	prev := uint(0)
	for _, h := range src {
		v := h - prev
		q, r := v/p, v%p

		// Write unary code of quotient
		if err := wr.Write(1<<(uint(q)+1)-2, int(q+1)); err != nil {
			return err
		}

		// Write remainder
		if err := wr.Write(r, bitLen); err != nil {
			return err
		}

		prev = h
	}

	return wr.Flush()
}
