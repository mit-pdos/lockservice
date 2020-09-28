package lockservice

//
// the lockservice Clerk lives in the client
// and maintains a little state.
//
type Clerk struct {
	primary string
	cid		uint64
	seq		uint64
}

func MakeClerk(primary string, cid uint64) *Clerk {
	ck := new(Clerk)
	ck.primary = primary
	ck.cid = cid
	ck.seq = 1
	return ck
}

//
// waits until the lock service grants us the lock
//
func (ck *Clerk) Lock(lockname uint64) bool {
    args := &LockArgs{Lockname:lockname, CID:ck.cid, Seq:ck.seq}
	ck.seq = ck.seq + 1

	// prepare the arguments.
	var reply LockReply

	// send an RPC request, wait for the reply.
	var ok = false
	for {
		ok = CallTryLock(ck.primary, args, &reply)
		if ok == true {
			if reply.OK { break }
			args.Seq = ck.seq
			ck.seq = ck.seq + 1
		}
	}
    return reply.OK
}

//
// ask the lock service to unlock a lock.
// returns true if the lock was previously held,
// false otherwise.
//

func (ck *Clerk) Unlock(lockname uint64) bool {
	// prepare the arguments.
    args := &UnlockArgs{Lockname:lockname, CID:ck.cid, Seq:ck.seq}
	ck.seq = ck.seq + 1

	var reply UnlockReply

	// send an RPC request, wait for the reply.
	var ok bool
	for {
		ok = CallUnlock(ck.primary, args, &reply)
		if ok == true {
		    break
		}
	}

	return reply.OK
}
