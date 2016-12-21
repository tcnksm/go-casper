package golombset

import (
	"bufio"
	"log"
	"os"
	"testing"
)

func BenchmarkGlombSet(b *testing.B) {
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !g.Search([]byte("alpha")) {
			b.Fatalf("expect not to be false")
		}
	}
}
