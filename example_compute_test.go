package compute_test

import (
	"fmt"

	"github.com/maxim2266/compute"
)

// summation function
func sum(args ...int) (r int) {
	for _, x := range args {
		r += x
	}

	return
}

// multiplication function
func mul(args ...int) (r int) {
	if len(args) > 0 {
		r = args[0]

		for _, x := range args[1:] {
			r *= x
		}
	}

	return
}

func Example() {
	// create a Pad
	pad := compute.NewPad[string, int]()

	// add values
	pad.SetVal("a", 1) // a = 1
	pad.SetVal("b", 2) // b = 2
	pad.SetVal("c", 3) // c = 3

	// add functions
	pad.SetFunc("x", sum, "a", "b") // x = a + b
	pad.SetFunc("y", sum, "b", "c") // y = b + c
	pad.SetFunc("z", mul, "x", "y") // z = x * y

	// calculate
	for k, v := range pad.Calc("x", "y", "z") {
		fmt.Printf("%s = %d\n", k, v)
	}

	// Output:
	// x = 3
	// y = 5
	// z = 15
}
