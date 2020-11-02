package lockservice

//
// the lockservice Clerk lives in the client
// and maintains a little state.
//
type Clerk struct {
	primary *LockServer
	cid     uint64
	seq     uint64
}

func MakeClerk(primary *LockServer, cid uint64) *Clerk {
	ck := new(Clerk)
	ck.primary = primary
	ck.cid = cid
	ck.seq = 1
	return ck
}

func (ck *Clerk) TryLock(lockname uint64) uint64 {
    overflow_guard_incr(ck.seq)
	// prepare the arguments.
	var args = &RPCRequest{Arg1: lockname, CID: ck.cid, Seq: ck.seq}
	ck.seq = ck.seq + 1

	// send an RPC request, wait for the reply.
	var errb = false
	reply := new(RPCReply)
	for {
		errb = CallRpc(ck.primary.TryLock, args, reply)
		if errb == false {
			break
		}
		continue
	}
	return reply.Ret
}

//
// ask the lock service to unlock a lock.
// returns true if the lock was previously held,
// false otherwise.
//
func (ck *Clerk) Unlock(lockname uint64) uint64 {
    overflow_guard_incr(ck.seq)
	// prepare the arguments.
	var args = &RPCRequest{Arg1: lockname, CID: ck.cid, Seq: ck.seq}
	ck.seq = ck.seq + 1

	// send an RPC request, wait for the reply.
	var errb = false
	reply := new(RPCReply)
	for {
		errb = CallRpc(ck.primary.Unlock, args, reply)
		if errb == false {
			break
		}
		continue
	}

	return reply.Ret
}

// Spins until we have the lock
func (ck *Clerk) Lock(lockname uint64) bool {
	for {
		if ck.TryLock(lockname) == 1 {
			break
		}
		continue
	}
	return true
}
