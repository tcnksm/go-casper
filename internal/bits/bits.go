package bits

import (
	"encoding/binary"
	"io"
)

// Writer
type Writer struct {
	n int  // current number of bits
	v uint // current accumulated value

	wr io.Writer
}

// NewWriter returns a new Writer.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		wr: w,
	}
}

// Write writes bits with give size n.
func (w *Writer) Write(bits uint, n int) error {
	w.v <<= uint(n)
	w.v |= bits & mask(n)
	w.n += n
	for w.n >= 8 {
		b := (w.v >> (uint(w.n) - 8)) & mask(8)
		if err := binary.Write(w.wr, binary.BigEndian, uint8(b)); err != nil {
			return err
		}
		w.n -= 8
	}
	w.v &= mask(8)

	return nil
}

// Flush writes any remaining bits to the underlying io.Writer.
// bits will be left-shifted.
func (w *Writer) Flush() error {
	if w.n != 0 {
		b := (w.v << (8 - uint(w.n))) & mask(8)
		if err := binary.Write(w.wr, binary.BigEndian, uint8(b)); err != nil {
			return err
		}
	}
	return nil
}

type Reader struct {
	n int  // current number of bits
	v uint // current accumulated value

	rd io.Reader
}

func (r *Reader) Read(n int) (uint, error) {
	for r.n <= n {
		r.v <<= 8
		var b uint8
		if err := binary.Read(r.rd, binary.BigEndian, &b); err != nil {
			return 0, err
		}
		r.v |= uint(b)

		r.n += 8
	}

	v := r.v >> uint(r.n-n)

	r.n -= n
	r.v &= mask(r.n)

	return v, nil
}

func mask(n int) uint {
	return (1 << uint(n)) - 1
}
