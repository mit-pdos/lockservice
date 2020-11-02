package lockservice

//
// RPC definitions for a simple lock service.
//

type RPCRequest struct {
	// Go's net/rpc requires that these field
	// names start with upper case letters!
	CID      uint64
	Seq      uint64
	Arg1 uint64
	Arg2 uint64
}
type RPCReply struct {
	Stale bool
	Ret uint64
}

// Call this before doing an increment that has risk of overflowing.
// If it's going to overflow, this'll loop forever, so the bad addition can never happen
func overflow_guard_incr(v uint64) {
	for v + 1 < v {
	}
}
