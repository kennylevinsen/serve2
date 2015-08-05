package proto

import (
	"testing"
)

func TestTLS(t *testing.T) {
	h := &TLS{}

	tests := []struct {
		payload  []byte
		match    bool
		required int
	}{
		{nil, false, 6},
		{[]byte{0x15}, false, 0},
		{[]byte{0x16, 0x02}, false, 0},
		{[]byte{0x16, 0x03}, false, 6},
		{[]byte{0x15, 0x03, 0x01, 0x00, 0xc4, 0x01}, false, 0},
		{[]byte{0x16, 0x02, 0x01, 0x00, 0xc4, 0x01}, false, 0},
		{[]byte{0x16, 0x02, 0x01, 0x00, 0xc4, 0x01}, false, 0},
		{[]byte{0x16, 0x03, 0x01, 0x00, 0xc4, 0x01}, true, 0},
		{[]byte{0x16, 0x03, 0x01, 0x00, 0x8d, 0x01}, true, 0},
	}

	for _, test := range tests {
		match, required := h.Check(test.payload, nil)
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
