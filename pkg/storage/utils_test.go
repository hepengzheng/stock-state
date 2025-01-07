package storage

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValueCodec(t *testing.T) {
	tester := func(value int32) {
		encoded := encodeValue(value)
		decoded, err := decodeValue(encoded)
		if err != nil {
			t.Fatalf("error decoding value: %v", err)
		}
		assert.Equal(t, value, decoded)
	}

	tester(2234)
	tester(0)
	tester(math.MaxInt32)
}

func TestValueComparison(t *testing.T) {
	tester := func(a, b int32) {
		if a > b {
			assert.True(t, encodeValue(a) > encodeValue(b))
		} else if a == b {
			assert.True(t, encodeValue(a) == encodeValue(b))
		} else {
			assert.True(t, encodeValue(b) > encodeValue(a))
		}
	}

	cases := []struct{ a, b int32 }{
		{1, 2}, {0, 10}, {1000000, 1},
		{math.MaxInt32, math.MaxInt32},
		{math.MaxInt32, math.MaxInt32 - 1},
	}
	for _, c := range cases {
		tester(c.a, c.b)
	}
}
