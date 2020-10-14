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

//
// waits until the lock service grants us the lock
//
func (ck *Clerk) Lock(lockname uint64) bool {
	var args = &LockArgs{Lockname: lockname, CID: ck.cid, Seq: ck.seq}
	ck.seq = ck.seq + 1

	// prepare the arguments.

	// send an RPC request, wait for the reply.
	var errb = false
	reply := new(LockReply)
	for {
		errb = CallTryLock(ck.primary, args, reply)
		if errb == false {
			if reply.OK {
				break
			}
			args = &LockArgs{Lockname: lockname, CID: ck.cid, Seq: ck.seq}
			ck.seq = ck.seq + 1
			continue
		}
		continue
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
	args := &UnlockArgs{Lockname: lockname, CID: ck.cid, Seq: ck.seq}
	ck.seq = ck.seq + 1

	// send an RPC request, wait for the reply.
	var errb bool
	reply := new(UnlockReply)
	for {
		errb = CallUnlock(ck.primary, args, reply)
		if errb == false {
			break
		}
		continue
	}

	return reply.OK
}
