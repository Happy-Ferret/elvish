package eval

import (
	"fmt"
	"strconv"
)

// Unwrappers are helper types for "unwrapping" values, the process for
// asserting certain properties of values and throwing exceptions when such
// properties are not satisfied.

type unwrapperInner struct {
	// ctx is the evaluation context.
	ctx *EvalCtx
	// description describes what is being unwrapped. It is used in error
	// messages.
	description string
	// begin and end contains positions in the source code to point to when
	// error occurs.
	begin, end int
	// values contain the Value's to unwrap.
	values []Value
}

func (u *unwrapperInner) error(want, gotfmt string, gotargs ...interface{}) {
	got := fmt.Sprintf(gotfmt, gotargs...)
	u.ctx.errorpf(u.begin, u.end, "%s must be %s; got %s", u.description,
		want, got)
}

// ValuesUnwrapper unwraps []Value.
type ValuesUnwrapper struct{ *unwrapperInner }

// Unwrap creates an Unwrapper.
func (ctx *EvalCtx) Unwrap(desc string, begin, end int, vs []Value) ValuesUnwrapper {
	return ValuesUnwrapper{&unwrapperInner{ctx, desc, begin, end, vs}}
}

// ExecAndUnwrap executes a ValuesOp and creates an Unwrapper for the obtained
// values.
func (ctx *EvalCtx) ExecAndUnwrap(desc string, op ValuesOp) ValuesUnwrapper {
	return ctx.Unwrap(desc, op.Begin, op.End, op.Exec(ctx))
}

// One unwraps the value to be exactly one value.
func (u ValuesUnwrapper) One() ValueUnwrapper {
	if len(u.values) != 1 {
		u.error("a single value", "%d values", len(u.values))
	}
	return ValueUnwrapper{u.unwrapperInner}
}

// ValueUnwrapper unwraps one Value.
type ValueUnwrapper struct{ *unwrapperInner }

func (u ValueUnwrapper) Any() Value {
	return u.values[0]
}

func (u ValueUnwrapper) String() String {
	s, ok := u.values[0].(String)
	if !ok {
		u.error("string", "%s", u.values[0].Kind())
	}
	return s
}

func (u ValueUnwrapper) Int() int {
	s := u.String()
	i, err := strconv.Atoi(string(s))
	if err != nil {
		u.error("integer", "%s", s)
	}
	return i
}

func (u ValueUnwrapper) NonNegativeInt() int {
	i := u.Int()
	if i < 0 {
		u.error("non-negative int", "%d", i)
	}
	return i
}

func (u ValueUnwrapper) FdOrClose() int {
	s := string(u.String())
	if s == "-" {
		return -1
	}
	return u.NonNegativeInt()
}

func (u ValueUnwrapper) Callable() Callable {
	c, ok := u.values[0].(Callable)
	if !ok {
		u.error("callable", "%s", u.values[0].Kind())
	}
	return c
}

func (u ValueUnwrapper) Iterable() Iterable {
	it, ok := u.values[0].(Iterable)
	if !ok {
		u.error("iterable", "%s", u.values[0].Kind())
	}
	return it
}
