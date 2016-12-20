package unary

import "testing"

func TestEncode(t *testing.T) {
	cases := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "10"},
		{2, "110"},
		{3, "1110"},
		{4, "11110"},
		{5, "111110"},
		{6, "1111110"},
		{7, "11111110"},
		{8, "111111110"},
		{9, "1111111110"},
	}

	for _, tc := range cases {
		if got := string(encode(tc.input)); got != tc.want {
			t.Errorf("Encode(%d)=%s, want=%s", tc.input, got, tc.want)
		}
	}
}

func TestDecode(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"0", 0},
		{"10", 1},
		{"110", 2},
		{"1110", 3},
		{"11110", 4},
		{"111110", 5},
		{"1111110", 6},
		{"11111110", 7},
		{"111111110", 8},
		{"1111111110", 9},
	}

	for _, tc := range cases {
		if got := decode([]byte(tc.input)); got != tc.want {
			t.Errorf("Decode(%d)=%s, want=%s", tc.input, got, tc.want)
		}
	}
}
