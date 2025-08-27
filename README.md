## compute: a generic streaming calculator.

[![GoDoc](https://godoc.org/github.com/maxim2266/compute?status.svg)](https://godoc.org/github.com/maxim2266/compute)
[![Go Report Card](https://goreportcard.com/badge/github.com/maxim2266/compute)](https://goreportcard.com/report/github.com/maxim2266/compute)
[![License: BSD 3-Clause](https://img.shields.io/badge/License-BSD_3--Clause-yellow.svg)](https://opensource.org/licenses/BSD-3-Clause)

Package `compute` implements a generic streaming calculator. The core type of the package is
`Pad` - a map from type `K cmp.Ordered` to either type `V any` or a function of type `func(...V) V`.
`Pad` provides methods for inserting keys and values/functions, and an iterator method that
takes a sequence of keys and produces a sequence of `K, V` pairs where each `V` is either an
existing value under the key `K`, or a result of calling the function under that key. During
the iteration it is guaranteed that each function is called at most once (aka lazy evaluation).

Example:
```Go
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
```

Possible applications:
* With `string` keys and values of some numeric type it becomes an expression evaluator with symbol table
  (as in the example above);
* With integer keys it can be used as an engine for spreadsheet-like calculations;
* With `string` values and appropriate functions one can even develop a `make`-like utility on top of it.

For API details see [documentation](https://godoc.org/github.com/maxim2266/compute).

#### Details
The key sequence for calculations does not have to be in any particular order, `Pad` handles all the
dependencies internally.

During the calculation existing values and functions in the `Pad` must not be modified, but
new values and functions can be added. Provided that the sequence of keys being iterated over also
includes those new values, a highly dynamic calculations can be performed.

The calculation itself does not modify the contents of the `Pad`, making it immutable and `panic`-safe,
but this also means that all intermediate results (i.e., the calculated values) get discarded when
the iteration ends.

The calculator stops with an error upon detecting a missing key or a cycle in the calculation graph.

Insertion of a new value/function is a rather cheap operation, and one `Pad` can potentially contain
millions of values.
