package compute_test

import (
	"fmt"

	"github.com/maxim2266/compute"
)

func add(a, b int) int {
	return a + b
}

func mul(a, b int) int {
	return a * b
}

func Example() {
	// create a Pad
	pad := compute.NewPad[string, int]()

	// add values
	pad.SetVal("a", 1) // a = 1
	pad.SetVal("b", 2) // b = 2
	pad.SetVal("c", 3) // c = 3

	// add functions
	pad.SetFunc2("x", add, "a", "b") // x = a + b
	pad.SetFunc2("y", add, "b", "c") // y = b + c
	pad.SetFunc2("z", mul, "x", "y") // z = x * y

	// calculate
	for k, v := range pad.Calc("x", "y", "z") {
		fmt.Printf("%s = %d\n", k, v)
	}

	// Output:
	// x = 3
	// y = 5
	// z = 15
}
