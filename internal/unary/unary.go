// package unary implements unary coding encode/decode function.
//
// https://en.wikipedia.org/wiki/Unary_coding
package unary

import "fmt"

// Encode encodes int to unary code.
func Encode(n int) []byte {
	return []byte(fmt.Sprintf("%b", 1<<(uint(n)+1)-2))
}

// Decode decodes unary coded bytes to int.
func Decode(b []byte) int {
	return len(b) - 1
}
