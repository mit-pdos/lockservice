package lockservice

//
// the lockservice Clerk lives in the client
// and maintains a little state.
//
type Clerk struct {
	primary string
	backup  string
	cid     uint64
	seq     uint64
}

func MakeClerk(primary string, backup string, cid uint64) *Clerk {
	ck := new(Clerk)
	ck.primary = primary
	ck.backup = backup
	ck.cid = cid
	ck.seq = 1
	return ck
}

//
// ask the lock service for a lock.
// returns true if the lock service
// granted the lock, false otherwise.
//
// you will have to modify this function.
//
func (ck *Clerk) Lock(lockname uint64) bool {
	args := &LockArgs{}
	args.Lockname = lockname
	args.CID = ck.cid
	args.Seq = ck.seq
	ck.seq++

	// prepare the arguments.
	var reply LockReply

	// send an RPC request, wait for the reply.
	var ok bool
	ok = CallLock(ck.primary, args, &reply)
	if ok == true {
		return reply.OK
	}

	ok = CallLock(ck.backup, args, &reply)
	if ok == true {
		return reply.OK
	}

	return false
}

//
// ask the lock service to unlock a lock.
// returns true if the lock was previously held,
// false otherwise.
//

func (ck *Clerk) Unlock(lockname uint64) bool {
	// prepare the arguments.
	args := &UnlockArgs{}
	args.Lockname = lockname
	args.CID = ck.cid
	args.Seq = ck.seq
	ck.seq++

	var reply UnlockReply

	// send an RPC request, wait for the reply.
	var ok bool
	ok = CallUnlock(ck.primary, args, &reply)
	if ok == true {
		return reply.OK
	}

	ok = CallUnlock(ck.backup, args, &reply)
	if ok == true {
		return reply.OK
	}

	return false
}
