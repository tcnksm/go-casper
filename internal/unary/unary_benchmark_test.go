package unary

import "testing"

func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Encode(10000)
	}
}
