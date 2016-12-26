package golombset

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"os"
	"sort"
	"strconv"
	"testing"
)

func TestScanner(t *testing.T) {
	cases := []struct {
		input []byte
		want  []uint
	}{
		{
			[]byte{0x97}, // 10010111
			[]uint{87},
		},
		{
			[]byte{0xcb, 0x80}, // 11001011 10000000
			[]uint{151, 0},
		},
	}

	for _, tc := range cases {
		sc := NewScanner(bytes.NewReader(tc.input), 26, 64)
		var index int
		for sc.Scan() {
			if got := sc.Value(); got != tc.want[index] {
				t.Fatalf("Scan=%d, want=%d", got, tc.want[index])
			}
			index++
		}

		if err := sc.Err(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEncoding(t *testing.T) {
	cases := []struct {
		input []uint
		want  []byte
	}{
		{
			[]uint{151},
			[]byte{0xcb, 0x80}, // 11001011 10000000
		},
	}

	for _, tc := range cases {
		var buf bytes.Buffer
		encoder := NewEncoder(&buf, 26, 64)
		if err := encoder.Encode(tc.input); err != nil {
			t.Fatal(err)
		}

		if got := buf.Bytes(); !bytes.Equal(got, tc.want) {
			t.Errorf("Encode=%x, want=%x", got, tc.want)
		}
	}
}

// This test used data set from [1].
//
// [1]: https://github.com/rasky/gcs
func TestGlombSet(t *testing.T) {

	n, p := uint(26), uint(64)
	file := "./testdata/words.nato"
	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	hashFunc := func(v []byte) uint {
		h := md5.New()
		h.Write(v)
		b := h.Sum(nil)

		s := hex.EncodeToString(b[12:16])
		i, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			panic(err)
		}
		return uint(i) % (n * p)
	}

	sc := bufio.NewScanner(f)
	a := make([]uint, 0, n)
	for sc.Scan() {
		a = append(a, hashFunc(sc.Bytes()))
	}

	sort.Slice(a, func(i, j int) bool {
		return a[i] < a[j]
	})

	diffs := make([]uint, 0, len(a))
	for i := 0; i < len(a); i++ {
		if i == 0 {
			diffs = append(diffs, a[i])
			continue
		}
		diffs = append(diffs, a[i]-a[i-1])
	}

	// Encode hash value array to Golomb-coded sets
	// and write it to buffer.
	var buf bytes.Buffer
	encoder := NewEncoder(&buf, n, p)
	if err := encoder.Encode(diffs); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		input string
		want  bool
	}{
		{"alpha", true},
		{"hotel", true},
		{"whiskey", true},
		{"november", true},

		{"Taichi", false},
		{"Nakashima", false},
	}

	for _, tc := range cases {
		sc := NewScanner(bytes.NewReader(buf.Bytes()), n, p)
		if got := Search(sc, hashFunc([]byte(tc.input))); got != tc.want {
			t.Errorf("Search(%s)=%t, want=%t", tc.input, got, tc.want)
		}
	}
}
