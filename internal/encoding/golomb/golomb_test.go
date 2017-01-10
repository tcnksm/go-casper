package golomb

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/tcnksm/go-casper/internal/bits"
)

func TestDecode(t *testing.T) {
	cases := []struct {
		input []byte
		p     uint
		want  uint
		err   error
	}{
		{
			[]byte{0xcb, 0x80}, // 11001011 10000000
			1 << 6,
			151,
			nil,
		},
		{
			[]byte{0xcb}, // 11001011
			1 << 5,
			75,
			io.EOF,
		},
		{
			[]byte{0x00}, // 00000000
			1 << 7,
			0,
			errPadding,
		},
	}

	for _, tc := range cases {
		rd := bytes.NewReader(tc.input)
		br := bits.NewReader(rd)
		got, err := decode(br, tc.p)
		if err != tc.err {
			t.Errorf("error=%v, want=%v", err, tc.err)
		}

		if got != tc.want {
			t.Errorf("decode=%v, want=%v", got, tc.want)
		}

	}
}

func TestDecodeAll(t *testing.T) {
	cases := []struct {
		input []byte
		p     uint
		want  []uint
	}{
		{
			[]byte{0xcb, 0x80}, // 11001011 10000000
			1 << 6,
			[]uint{151},
		},

		{
			[]byte{0xcb, 0xcf}, // 11001011 11001111
			1 << 5,
			[]uint{75, 154},
		},
	}

	for _, tc := range cases {
		got, err := DecodeAll(tc.input, tc.p)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Decode=%v, want=%v", got, tc.want)
		}

	}
}

func TestEncoding(t *testing.T) {
	cases := []struct {
		input []uint
		p     uint
		want  []byte
	}{
		{
			[]uint{151},
			1 << 6,
			[]byte{0xcb, 0x80}, // 11001011 10000000
		},
	}

	for _, tc := range cases {
		var buf bytes.Buffer
		if err := Encode(&buf, tc.input, tc.p); err != nil {
			t.Fatal(err)
		}

		if got := buf.Bytes(); !bytes.Equal(got, tc.want) {
			t.Errorf("Encode=%x, want=%x", got, tc.want)
		}
	}
}

func ExampleGlombSet() {
	// Number of elements and false positive probability.
	//
	// Minimum number of bits is N*log(P)
	// = 26 * log(1<<6) = 156 bits = 19 bytes
	N, P := uint(26), uint(1<<6)

	// Example data set comes from https://github.com/rasky/gcs
	file := "./testdata/words.nato"
	f, _ := os.Open(file)
	sc := bufio.NewScanner(f)
	defer f.Close()

	hashF := func(v []byte) uint {
		h := md5.New()
		h.Write(v)
		b := h.Sum(nil)

		s := hex.EncodeToString(b[12:16])
		i, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			panic(err)
		}
		return uint(i) % (N * P)
	}

	a := make([]uint, 0, N)
	for sc.Scan() {
		a = append(a, hashF(sc.Bytes()))
	}
	sort.Slice(a, func(i, j int) bool {
		return a[i] < a[j]
	})

	// Encode hash value array to Golomb-coded sets
	// and write it to buffer.
	var buf bytes.Buffer
	Encode(&buf, a, P)

	fmt.Printf("%x", buf.Bytes())
	// Output: cba920f780663a061f2065198ab1032d624c50331e66ae9818
}
