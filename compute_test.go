package compute

import (
	"fmt"
	"slices"
	"testing"
)

func TestSimple(t *testing.T) {
	defer clearTrace()

	pad := NewPadFrom(basePad)

	if n := pad.Size(); n != 7 {
		t.Fatalf("unexpected Pad size: %d", n)
	}

	// check values
	i := 1

	for k, v := range pad.Calc(1, 2, 3, 4) {
		if i > 4 {
			t.Fatal("unexpected iteration")
		}

		if k != i {
			t.Fatalf("unexpected key: %d instead of %d", k, i)
		}

		if v != i {
			t.Fatalf("unexpected value (key %d): %d instead of %d", i, v, i)
		}

		i++
	}

	// calculations
	res, err := calcInts(pad, 5, 6, 7)

	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(res, []int{3, 7, 10}) {
		t.Fatalf("unexpected result: %v", res)
	}

	if !slices.Equal(trace, []int{5, 6, 7}) {
		t.Fatalf("unexpected evaluation trace: %v", trace)
	}

	clearTrace()

	// test Clear() function
	pad.Clear()
	pad.UpdateFrom(basePad)

	// calculations, in a different order
	if res, err = calcInts(pad, 7, 6, 5); err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(res, []int{10, 7, 3}) {
		t.Fatalf("unexpected result: %v", res)
	}

	if !slices.Equal(trace, []int{5, 6, 7}) {
		t.Fatalf("unexpected evaluation trace: %v", trace)
	}
}

func TestMissingKey(t *testing.T) {
	defer clearTrace()

	pad := NewPadFrom(basePad)

	// missing value
	for k, v := range pad.Calc(123) {
		t.Fatalf("unexpected iteration: %d, %d", k, v)
	}

	if pad.Err == nil || pad.Err.Error() != `missing key "123"` {
		t.Fatalf("unexpected error: %v", pad.Err)
	}

	// missing function parameter
	pad.SetFunc(7, wrap(sum, 7), 5, 8) // 8 is missing

	res, err := calcInts(pad, 5, 6, 7)

	if err == nil || err.Error() != `missing key "8"` {
		t.Fatalf("unexpected error: %v", err)
	}

	if !slices.Equal(res, []int{3, 7}) {
		t.Fatalf("unexpected result: %v", res)
	}

	if !slices.Equal(trace, []int{5, 6}) {
		t.Fatalf("unexpected evaluation trace: %v", trace)
	}

	// another missing function parameter
	pad.SetFunc(7, wrap(sum, 7), 5, 6)
	pad.Delete(1)

	clearTrace()

	if res, err = calcInts(pad, 5, 6, 7); err == nil {
		t.Fatal("missing error")
	}

	if err == nil || err.Error() != `missing key "1"` {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res) > 0 {
		t.Fatalf("unexpected result: %v", res)
	}

	if len(trace) > 0 {
		t.Fatalf("unexpected evaluation trace: %v", trace)
	}
}

func TestCycle(t *testing.T) {
	defer clearTrace()

	pad := NewPadFrom(basePad)

	pad.SetFunc(5, wrap(sum, 5), 1, 7) // 5 -> 7 -> 5

	res, err := calcInts(pad, 5, 6, 7)

	if err == nil || err.Error() != `cycle detected on key "5"` {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res) > 0 {
		t.Fatalf("unexpected result: %v", res)
	}

	if len(trace) > 0 {
		t.Fatalf("unexpected evaluation trace: %v", trace)
	}
}

func BenchmarkCompute(b *testing.B) {
	pad := NewPad[int, int]()

	pad.SetVal(1, 1)
	pad.SetVal(2, 2)
	pad.SetVal(3, 3)

	pad.SetFunc2(4, add, 1, 2)
	pad.SetFunc2(5, add, 2, 3)
	pad.SetFunc2(6, mul, 4, 5)

	for b.Loop() {
		for k, v := range pad.Calc(6) { // one iteration only
			if k != 6 || v != 15 {
				b.Fatalf("invalid result: %d, %d", k, v)
			}
		}
	}
}

func calcInts(pad *Pad[int, int], keys ...int) (res []int, err error) {
	i := 0

	for k, v := range pad.Calc(keys...) {
		if i == len(keys) {
			err = fmt.Errorf("[%d]: unexpected iteration", i)
			return
		}

		if k != keys[i] {
			err = fmt.Errorf("[%d]: unexpected key: %d instead of %d", i, k, keys[i])
			return
		}

		res = append(res, v)
		i++
	}

	err = pad.Err
	return
}

var basePad = NewPad[int, int]()
var trace []int

func clearTrace() {
	trace = trace[:0]
}

func init() {
	basePad.SetVal(1, 1)
	basePad.SetVal(2, 2)
	basePad.SetVal(3, 3)
	basePad.SetVal(4, 4)

	basePad.SetFunc(5, wrap(sum, 5), 1, 2)
	basePad.SetFunc(6, wrap(sum, 6), 3, 4)
	basePad.SetFunc(7, wrap(sum, 7), 5, 6)
}

func wrap(fn func(...int) int, key int) func(...int) int {
	return func(args ...int) int {
		trace = append(trace, key)
		return fn(args...)
	}
}

func sum(args ...int) (res int) {
	for _, x := range args {
		res += x
	}

	return
}

func add(a, b int) int {
	return a + b
}

func mul(a, b int) int {
	return a * b
}
