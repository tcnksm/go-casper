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

type Reader struct{}

func mask(n int) uint {
	return (1 << uint(n)) - 1
}
