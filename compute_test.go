package compute

import (
	"fmt"
	"testing"
)

func TestSimple(t *testing.T) {
	var pad Pad[int, int]

	pad.SetVal(1, 1)
	pad.SetVal(2, 2)
	pad.SetVal(3, 3)
	pad.SetVal(4, 4)

	pad.SetFunc(5, sum, 1, 2)
	pad.SetFunc(6, sum, 3, 4)
	pad.SetFunc(7, sum, 5, 6)

	if err := calcInts(&pad, []int{3, 7, 10}, 5, 6, 7); err != nil {
		t.Fatal(err)
	}

	if err := calcInts(&pad, []int{10, 7, 3}, 7, 6, 5); err != nil {
		t.Fatal(err)
	}
}

func calcInts(pad *Pad[int, int], values []int, keys ...int) error {
	i := 0

	for k, v := range pad.Calc(keys...) {
		if k != keys[i] {
			return fmt.Errorf("[%d]: unexpected key: %d instead of %d", i, k, keys[i])
		}

		if v != values[i] {
			return fmt.Errorf("[%d]: unexpected value (key %d): %d instead of %d", i, k, v, values[i])
		}

		i++
	}

	if i != len(values) {
		return fmt.Errorf("unexpected number of iterations: %d instead of %d", i, len(values))
	}

	return nil
}

func sum(args ...int) (res int) {
	for _, x := range args {
		res += x
	}

	return
}
