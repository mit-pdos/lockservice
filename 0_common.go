package lockservice

import (
	"github.com/mit-pdos/lockservice/grove_ffi"
)

type HostName = grove_ffi.HostName

func nondet() bool {
	return true
}

// Call this before doing an increment that has risk of overflowing.
// If it's going to overflow, this'll loop forever, so the bad addition can never happen
func overflow_guard_incr(v uint64) {
	for v+1 < v {
	}
}
