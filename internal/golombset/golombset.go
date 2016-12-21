package golombset

import (
	"bytes"
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

func (g *GolombSet) Search(p []byte) bool {
	hash := g.hash(p)
	var prev uint
	for _, v := range g.Codes {
		idx := bytes.Index(v, []byte("0"))
		quot := unary.Decode(v[0 : idx+1])
		rem, _ := strconv.ParseInt(string(v[idx+1:]), 2, 64)
		o := g.P*uint(quot) + uint(rem)

		h := o + prev
		if h == hash {
			return true
		}
		prev = h
	}
	return false
}

func (g *GolombSet) Encode() error {
	// Sort hash values
	sort.Slice(g.Store, func(i, j int) bool {
		return g.Store[i] < g.Store[j]
	})

	diffs := make([]int, 0, len(g.Store))
	for i := 0; i < len(g.Store); i++ {
		if i == 0 {
			diffs = append(diffs, int(g.Store[i]))
			continue
		}
		diffs = append(diffs, int(g.Store[i]-g.Store[i-1]))
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
