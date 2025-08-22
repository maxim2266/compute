package compute

import (
	"cmp"
	"fmt"
	"iter"
	"maps"
	"slices"
)

//go:generate tools/gen-functions functions.go functions_test.go

const minPadSize = 10

type Pad[K cmp.Ordered, V any] struct {
	env map[K]any // V | *formula[K, V]
	Err error     // computation error, if any
}

func NewPad[K cmp.Ordered, V any]() *Pad[K, V] {
	return NewPadSize[K, V](0)
}

func NewPadSize[K cmp.Ordered, V any](size int) *Pad[K, V] {
	return &Pad[K, V]{env: make(map[K]any, max(minPadSize, size))}
}

func NewPadFrom[K cmp.Ordered, V any](src *Pad[K, V]) *Pad[K, V] {
	return NewPadSize[K, V](len(src.env)).UpdateFrom(src)
}

func (p *Pad[K, V]) UpdateFrom(src *Pad[K, V]) *Pad[K, V] {
	maps.Copy(p.env, src.env)
	return p
}

func (p *Pad[K, V]) SetVal(key K, val V) {
	p.env[key] = val
}

func (p *Pad[K, V]) SetFunc(key K, fn func(...V) V, args ...K) {
	p.env[key] = &formula[K, V]{fn: fn, args: args}
}

func (p *Pad[K, V]) Delete(key K) {
	delete(p.env, key)
}

func (p *Pad[K, V]) Size() int {
	return len(p.env)
}

func (p *Pad[K, V]) Calc(keys ...K) iter.Seq2[K, V] {
	return p.CalcSeq(slices.Values(keys))
}

func (p *Pad[K, V]) CalcSeq(keys iter.Seq[K]) iter.Seq2[K, V] {
	// iterator function
	return func(yield func(K, V) bool) {
		// clear previous error
		p.Err = nil

		// calculator
		calc := calculator[K, V]{
			values: make(map[K]V),
			active: make(map[K]struct{}),
		}

		// iteration
		for key := range keys {
			// check what we've got under this key
			switch x := p.env[key].(type) {
			case nil: // nothing
				p.Err = fmt.Errorf(`missing key "%v"`, key)
				return

			case V: // value
				if !yield(key, x) {
					return
				}

			case *formula[K, V]: // formula to calculate
				// check if there is a value for it
				val, ok := calc.values[key]

				if !ok {
					// calculate the formula
					if val, p.Err = calc.eval(p.env, key, x); p.Err != nil {
						return
					}
				}

				// yield the computed value
				if !yield(key, val) {
					return
				}

			default: // must never happen
				panic("compute.Pad: invalid cell type")
			}
		}
	}
}

// calculator
type calculator[K cmp.Ordered, V any] struct {
	values map[K]V             // computed values
	stack  []computation[K, V] // stack of computations
	active map[K]struct{}      // cycle detector
}

func (calc *calculator[K, V]) push(key K, form *formula[K, V]) error {
	if _, yes := calc.active[key]; yes {
		return fmt.Errorf(`cycle detected on key "%v"`, key)
	}

	calc.stack = append(calc.stack, computation[K, V]{key: key, form: form})
	calc.active[key] = struct{}{}

	return nil
}

func (calc *calculator[K, V]) pop() bool {
	i := len(calc.stack)

	if i > 0 {
		i--
		delete(calc.active, calc.stack[i].key)
		calc.stack = calc.stack[:i]
	}

	return i == 0
}

func (calc *calculator[K, V]) eval(env map[K]any, key K, form *formula[K, V]) (value V, err error) {
	if err = calc.push(key, form); err != nil {
		return
	}

loop:
	for {
		// computation object
		c := &calc.stack[len(calc.stack)-1] // WARN: `c` is only valid till the next push

		// compute arguments
		for _, k := range c.form.args[len(c.args):] {
			// check values
			if val, ok := calc.values[k]; ok {
				c.args = append(c.args, val)
			} else {
				// check environment
				switch x := env[k].(type) {
				case nil:
					// not found
					err = fmt.Errorf(`missing key "%v"`, k)
					return

				case V:
					// it's a value
					c.args = append(c.args, x)

				case *formula[K, V]:
					// it's a formula to calculate
					if err = calc.push(k, x); err != nil {
						return
					}

					continue loop

				default: // must never happen
					panic("compute.Pad: invalid cell type")
				}
			}
		}

		// calculate the formula
		value = c.form.fn(c.args...)
		calc.values[c.key] = value

		// return when stack is empty
		if calc.pop() {
			return
		}
	}
}

// formula
type formula[K cmp.Ordered, V any] struct {
	fn   func(...V) V
	args []K
}

// computation
type computation[K cmp.Ordered, V any] struct {
	key  K
	form *formula[K, V]
	args []V
}
