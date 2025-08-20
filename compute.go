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
		// calculator
		calc := calculator[K, V]{
			values: make(map[K]V),
			stack:  evalStack[K, V]{active: make(map[K]struct{})},
		}

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
					calc.eval(p.env, key, c)

					if p.Err != nil {
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
	values map[K]V         // computed values
	stack  evalStack[K, V] // stack of computation[K, V]
}

func (calc *calculator[K, V]) eval(env map[K]any, key K, form *formula[K, V]) (err error) {
	if err = calc.stack.push(key, form); err != nil {
		return
	}

loop:
	for {
		// computation object
		c := calc.stack.top() // WARN: `c` is only valid till the next stack.push

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
					return errors.New("compute.calc: missing key")

				case V:
					// it's a value
					c.args = append(c.args, val)

				case *formula[K, V]:
					// it's a formula to calculate
					if err = calc.stack.push(key, val); err != nil {
						return
					}

					continue loop

				default: // must never happen
					panic("invalid arg type in p.env")
				}
			}
		}

		// calculate the formula
		calc.values[c.key] = c.form.fn(c.args...)

		// return when stack is empty
		if !calc.stack.pop() {
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

// stack
type evalStack[K comparable, V any] struct {
	stack  []computation[K, V]
	active map[K]struct{}
}

func (s *evalStack[K, V]) push(key K, form *formula[K, V]) error {
	if _, yes := s.active[key]; yes {
		return errors.New("cycle")
	}

	s.stack = append(s.stack, computation[K, V]{key: key, form: form})
	s.active[key] = struct{}{}

	return nil
}

func (s *evalStack[K, V]) pop() bool {
	if i := len(s.stack); i > 0 {
		i--
		delete(s.active, s.stack[i].key)
		s.stack = s.stack[:i]
	}

	return len(s.stack) > 0
}

func (s *evalStack[K, V]) top() *computation[K, V] {
	if i := len(s.stack); i > 0 {
		return &s.stack[i-1] // WARN: the arrdess is valid only till the next push()
	}

	panic("compute.top: empty evaluation stack") // must never happen
}
