package crash_barrier

import (
	ffi "../grove_ffi"
	"github.com/tchajed/marshal"
	"sync"
)

const Filename = "crash_barrier_example_counter"

var mu sync.Mutex // synchronize access to the counter

// id is a unique identifier for this invocation of incr_unsafe.  After crash
// and recover, we should be able toc all incr_safe with the same id, and it
// should have the same effect as it would have before crashing

func get() uint64 {
	return marshal.NewDec(ffi.Read(Filename)).GetInt()
}

func put(v uint64) {
	e := marshal.NewEnc(8)
	e.PutInt(v)
	ffi.Write(Filename, e.Finish())
}

func incr_unsafe(id uint64) uint64 {
	mu.Lock()
	v := get()
	v = v + 1
	put(v)
	mu.Unlock()
	return v
}

type State struct {
	id  uint64
	tmp uint64
	pc  uint64
}

func crash_op_start(id uint64) State {
	s := State{id: id, pc: 0}
	filename := "state_" + ffi.U64ToString(s.id)
	c := ffi.Read(filename)
	if len(c) == 0 {
		return s
	}
	d := marshal.NewDec(c)
	s.pc = d.GetInt()
	s.tmp = d.GetInt()
	return s
}

func (s *State) crash_barrier(pc uint64) {
	filename := "state_" + ffi.U64ToString(s.id)
	s.pc = pc
	e := marshal.NewEnc(16)
	e.PutInt(s.pc)
	e.PutInt(s.tmp)
	ffi.Write(filename, e.Finish())
}

func incr_safe(id uint64) uint64 {
	mu.Lock()
	// begin boilerplate
	c := crash_op_start(id)
	if c.pc == 1 {
		goto bar1
	}
	// end of boilerplate

	c.tmp = get()
	c.tmp = c.tmp + 1

bar1:
	c.crash_barrier(1)

	put(c.tmp)
	mu.Unlock()
	return c.tmp
}

// the goal of this crash barrier thingy is this:
// we can reason about the two lines up to the barrier, and come up with some
// idempotent spec for them
//
// Then, we can come up with a separate idempotent spec for the part after the
// crash barrier where the precondition is the postcondition of the first part.
// The postcondition of the first part/precondition of the second part is a
// predicate on c, so any relevant state should put there. One easy way to deal
// with that would be to have an array tmps integers in the State type that
// gets serialized and deserialized. The tmps array would be what gets saved at
// barriers.
//
// It's not necessarily valuable to literally write code like this, but this
// might give us some generic structure and reasoning principles.
