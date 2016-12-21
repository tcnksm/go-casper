package unary

import "testing"

func TestEncodeDecode(t *testing.T) {
	cases := []struct {
		i int
		u string
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
		u := Encode(tc.i)
		if got, want := string(u), tc.u; got != want {
			t.Errorf("Encode(%d)=%s, want=%s", tc.i, got, want)
		}

		i := Decode(u)
		if got, want := i, tc.i; got != want {
			t.Errorf("Decode(%s)=%d, want=%d", string(u), got, want)
		}
	}
}
