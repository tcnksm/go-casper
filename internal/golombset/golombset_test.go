package golombset

import (
	"bufio"
	"log"
	"os"
	"testing"
)

func TestGlombSet(t *testing.T) {
	file := "./testdata/words.nato"
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	g := &GolombSet{
		N: 26,
		P: 64,
	}

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		g.Add(sc.Bytes())
	}
	g.Encode()

	cases := []struct {
		input string
		want  bool
	}{
		{"alpha", true},
		{"hotel", true},
		{"whiskey", true},
		{"november", true},

		{"pen", false},
		{"apple", false},
		{"taichi", false},
		{"nakashima", false},
	}

	for _, tc := range cases {
		if got := g.Search([]byte(tc.input)); got != tc.want {
			t.Errorf("Search(%s)=%t, want=%t", tc.input, got, tc.want)
		}
	}
}
