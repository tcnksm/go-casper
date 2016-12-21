package golombset

import (
	"hash/fnv"
	"sort"
	"strconv"

	"github.com/tcnksm/go-casper/internal/unary"
)

type GolombSet struct {
	N uint // number of elements
	P uint // false-positive probability

	Store []uint
	Diffs []int
	Codes [][]byte
}

func (g *GolombSet) Add(p []byte) error {
	g.Store = append(g.Store, g.hash(p))
	return nil
}

func (g *GolombSet) Encode() error {
	// Sort hash values
	sort.Slice(g.Store, func(i, j int) bool {
		return g.Store[i] < g.Store[j]
	})

	diffs := make([]int, 0, len(g.Store))
	for i := 0; i < len(g.Store)-1; i++ {
		diffs = append(diffs, int(g.Store[i+1]-g.Store[i]))
	}
	g.Diffs = diffs

	for _, v := range diffs {
		quot, rem := uint(v)/g.P, uint(v)%g.P

		quotB := unary.Encode(int(quot))
		remB := []byte(strconv.FormatInt(int64(rem), 2))
		gv := append(quotB, remB...)
		g.Codes = append(g.Codes, gv)
	}
	return nil
}

func (g *GolombSet) hash(p []byte) uint {
	h := fnv.New64a()
	h.Write(p)
	i := uint(h.Sum64())
	return i % (g.N * g.P)
}
