package security

import "testing"

func TestZeroBytes(t *testing.T) {
	data := []byte{1, 2, 3}
	ZeroBytes(data)

	for i, value := range data {
		if value != 0 {
			t.Fatalf("byte %d was not zeroed", i)
		}
	}
}
