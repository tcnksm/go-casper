package bits

import (
	"bytes"
	"fmt"
	"testing"
)

func TestWriter(t *testing.T) {
	cases := []struct {
		size   int
		inputs []uint
		want   []byte
	}{
		{8, []uint{255}, []byte{0xff}},
		{4, []uint{15, 15}, []byte{0xff}},
		{2, []uint{3, 3, 3, 3}, []byte{0xff}},
		{1, []uint{1, 1, 1, 1, 1, 1, 1, 1}, []byte{0xff}},

		{4, []uint{15, 15, 15}, []byte{0xff, 0xf0}},
		{2, []uint{3, 3, 3, 3, 3, 3}, []byte{0xff, 0xf0}},
	}

	for _, tc := range cases {
		var buf bytes.Buffer
		writer := NewWriter(&buf)

		for _, input := range tc.inputs {
			if err := writer.Write(input, tc.size); err != nil {
				t.Fatalf("Write should not fail: %s", err)
			}
		}

		if err := writer.Flush(); err != nil {
			t.Fatalf("Flush should not fail: %s", err)
		}

		if !bytes.Equal(buf.Bytes(), tc.want) {
			t.Errorf("Write writes %x, want %x", buf.Bytes(), tc.want)
		}
	}
}

func TestMask(t *testing.T) {
	cases := []struct {
		input int
		want  string
	}{
		{8, "11111111"},
		{4, "00001111"},
		{2, "00000011"},
	}

	for _, tc := range cases {
		m := mask(tc.input)
		if got := fmt.Sprintf("%08b", m); got != tc.want {
			t.Errorf("mask(%d)=%s,want=%s", tc.input, got, tc.want)
		}
	}
}
