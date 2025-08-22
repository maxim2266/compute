package compute

import (
	"fmt"
	"iter"
	"maps"
	"slices"
)

const minPadSize = 10

type Pad[K comparable, V any] struct {
	env map[K]any // V | *formula[K, V]
	Err error     // computation error, if any
}

func NewPad[K comparable, V any]() *Pad[K, V] {
	return NewPadSize[K, V](0)
}

func NewPadSize[K comparable, V any](size int) *Pad[K, V] {
	return &Pad[K, V]{env: make(map[K]any, max(minPadSize, size))}
}

func NewPadFrom[K comparable, V any](src *Pad[K, V]) *Pad[K, V] {
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

func (p *Pad[K, V]) SetFunc0(key K, fn func() V) {
	p.env[key] = &formula[K, V]{fn: func(...V) V { return fn() }}
}

func (p *Pad[K, V]) SetFunc1(key K, fn func(V) V, arg K) {
	p.env[key] = &formula[K, V]{
		fn:   func(args ...V) V { return fn(args[0]) },
		args: []K{arg},
	}
}

func (p *Pad[K, V]) SetFunc2(key K, fn func(V, V) V, arg1, arg2 K) {
	p.env[key] = &formula[K, V]{
		fn:   func(args ...V) V { return fn(args[0], args[1]) },
		args: []K{arg1, arg2},
	}
}

func (p *Pad[K, V]) SetFunc3(key K, fn func(V, V, V) V, arg1, arg2, arg3 K) {
	p.env[key] = &formula[K, V]{
		fn:   func(args ...V) V { return fn(args[0], args[1], args[2]) },
		args: []K{arg1, arg2, arg3},
	}
}

func (p *Pad[K, V]) SetFunc4(key K, fn func(V, V, V, V) V, arg1, arg2, arg3, arg4 K) {
	p.env[key] = &formula[K, V]{
		fn:   func(args ...V) V { return fn(args[0], args[1], args[2], args[3]) },
		args: []K{arg1, arg2, arg3, arg4},
	}
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
			switch v := p.env[key].(type) {
			case nil: // nothing
				p.Err = fmt.Errorf(`missing key "%v"`, key)
				return

			case V: // value
				if !yield(key, v) {
					return
				}

			case *formula[K, V]: // formula to calculate
				// check if there is a value for it
				val, ok := calc.values[key]

				if !ok {
					// calculate the formula
					if val, p.Err = calc.eval(p.env, key, v); p.Err != nil {
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
type calculator[K comparable, V any] struct {
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
	if i := len(calc.stack); i > 0 {
		i--
		delete(calc.active, calc.stack[i].key)
		calc.stack = calc.stack[:i]
	}

	return len(calc.stack) > 0
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
		for _, key := range c.form.args[len(c.args):] {
			// check values
			if val, ok := calc.values[key]; ok {
				c.args = append(c.args, val)
			} else {
				// check environment
				switch val := env[key].(type) {
				case nil:
					// not found
					err = fmt.Errorf(`missing key "%v"`, key)
					return

				case V:
					// it's a value
					c.args = append(c.args, val)

				case *formula[K, V]:
					// it's a formula to calculate
					if err = calc.push(key, val); err != nil {
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
		if !calc.pop() {
			return
		}
	}
}

// formula
type formula[K comparable, V any] struct {
	fn   func(...V) V
	args []K
}

// computation
type computation[K comparable, V any] struct {
	key  K
	form *formula[K, V]
	args []V
}
