package lockservice

//
// RPC definitions for a simple lock service.
//

//
// TryLock(lockname) returns OK=true if the lock is not held.
// If it is held, it returns OK=false immediately.
//
type TryLockRequest struct {
	// Go's net/rpc requires that these field
	// names start with upper case letters!
	CID      uint64
	Seq      uint64
	Args uint64 // lock name
}

type TryLockReply struct {
	Stale bool
	Ret uint64
}

//
// Unlock(lockname) returns OK=true if the lock was held.
// It returns OK=false if the lock was not held.
//
type UnlockRequest struct {
	CID      uint64
	Seq      uint64
	Args uint64
}

type UnlockReply  = TryLockReply 

// Call this before doing an increment that has risk of overflowing.
// If it's going to overflow, this'll loop forever, so the bad addition can never happen
func overflow_guard_incr(v uint64) {
	for v + 1 < v {
	}
}
