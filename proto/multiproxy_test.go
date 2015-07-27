package proto

import (
	"testing"
)

func TestMultiProxy(t *testing.T) {
	methods := [][]byte{
		[]byte("GET"),
		[]byte("PUT"),
		[]byte("HEAD"),
		[]byte("POST"),
		[]byte("TRACE"),
		[]byte("PATCH"),
		[]byte("DELETE"),
		[]byte("OPTIONS"),
		[]byte("CONNECT"),
	}
	h := NewMultiProxy(methods, "", "")

	tests := []struct {
		payload  []byte
		match    bool
		required int
	}{
		{nil, false, 3},
		{[]byte("G"), false, 3},
		{[]byte("P"), false, 3},
		{[]byte("H"), false, 4},
		{[]byte("T"), false, 5},
		{[]byte("D"), false, 6},
		{[]byte("O"), false, 7},
		{[]byte("C"), false, 7},
		{[]byte("A"), false, 0},
		{[]byte("CO"), false, 7},
		{[]byte("GET"), true, 0},
		{[]byte("CONNACT"), false, 0},
		{[]byte("GET /index.html MultiProxy/1.1"), true, 0},
	}

	for _, test := range tests {
		match, required := h.Check(test.payload)
		if test.match != match {
			t.Errorf("match not correct for %q: was %t, expected %t",
				test.payload, match, test.match)
		}
		if test.required != required {
			t.Errorf("required not correct for %q: was %d, expected %d",
				test.payload, required, test.required)
		}
	}
}
