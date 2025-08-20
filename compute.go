package compute

import (
	"errors"
	"iter"
	"slices"
)

type Pad[K comparable, V any] struct {
	env map[K]any // V | *formula[K, V]
	Err error     // computation error, if any
}

func (p *Pad[K, V]) SetVal(key K, val V) {
	if p.env == nil {
		p.env = make(map[K]any)
	}

	p.env[key] = val
}

func (p *Pad[K, V]) SetFunc(key K, fn func(...V) V, args ...K) {
	if p.env == nil {
		p.env = make(map[K]any)
	}

	p.env[key] = &formula[K, V]{
		fn:   fn,
		args: args,
	}
}

func (p *Pad[K, V]) Delete(key K) {
	delete(p.env, key)
}

func (p *Pad[K, V]) Calc(args ...K) iter.Seq2[K, V] {
	return p.CalcSeq(slices.Values(args))
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
			switch c := p.env[key].(type) {
			case nil: // nothing
				p.Err = errors.New("missing key")
				return

			case V: // value
				if !yield(key, c) {
					return
				}

			case *formula[K, V]: // formula to calculate
				// check if there is a value for it
				val, ok := calc.values[key]

				if !ok {
					// calculate the formula
					if p.Err = calc.eval(p.env, key, c); p.Err != nil {
						return
					}

					val = calc.values[key]
				}

				// yield the computed value
				if !yield(key, val) {
					return
				}

			default: // must never happen
				panic("invalid cell type")
			}
		}
	}
}

// calculator
type calculator[K comparable, V any] struct {
	values map[K]V
	stack  []computation[K, V]
	active map[K]struct{}
}

func (c *calculator[K, V]) push(key K, form *formula[K, V]) error {
	if _, yes := c.active[key]; yes {
		return errors.New("cycle")
	}

	c.stack = append(c.stack, computation[K, V]{key: key, form: form})
	c.active[key] = struct{}{}

	return nil
}

func (c *calculator[K, V]) pop() bool {
	if i := len(c.stack); i > 0 {
		i--
		delete(c.active, c.stack[i].key)
		c.stack = c.stack[:i]
	}

	return len(c.stack) > 0
}

func (calc *calculator[K, V]) eval(env map[K]any, key K, form *formula[K, V]) (err error) {
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
					return errors.New("calculator.eval: missing key")

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
					panic("invalid arg type in env")
				}
			}
		}

		// calculate the formula
		calc.values[c.key] = c.form.fn(c.args...)

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
