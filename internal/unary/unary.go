// package unary implements unary coding encode/decode function.
//
// https://en.wikipedia.org/wiki/Unary_coding
package unary

// Encode encodes int to unary code.
// TODO(tcnksm): This is too naive implementation
func Encode(n int) []byte {
	buf := make([]byte, n+1)
	for i := 0; i < n; i++ {
		buf[i] = '1'
	}
	buf[n] = '0'

	return buf
}

// Decode decodes unary coded bytes to int.
// TODO(tcnksm): This is too naive implementation
func Decode(b []byte) int {
	return len(b) - 1
}
