package lockservice

//
// the lockservice Clerk lives in the client
// and maintains a little state.
//
type Clerk struct {
	primary *LockServer
	client *RPCClient
}

func MakeClerk(primary *LockServer, cid uint64) *Clerk {
	ck := new(Clerk)
	ck.primary = primary
	ck.client = MakeRPCClient(cid)
	return ck
}

func (ck *Clerk) TryLock(lockname uint64) bool {
	return ck.client.MakeRequest(ck.primary.TryLock, RPCVals{U64_1:lockname}) != 0
}

//
// ask the lock service to unlock a lock.
// returns true if the lock was previously held,
// false otherwise.
//
func (ck *Clerk) Unlock(lockname uint64) bool {
	return ck.client.MakeRequest(ck.primary.Unlock, RPCVals{U64_1:lockname}) != 0
}

// Spins until we have the lock
func (ck *Clerk) Lock(lockname uint64) bool {
	for {
		if ck.TryLock(lockname) {
			break
		}
		continue
	}
	return true
}
