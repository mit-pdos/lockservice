package lockservice

import (
	"github.com/mit-pdos/lockservice/grove_common"
)

//
// the lockservice Clerk lives in the client
// and maintains a little state.
//
type Clerk struct {
	host   uint64
	client *RPCClient
}

const LOCK_TRYLOCK uint64 = 1
const LOCK_UNLOCK uint64 = 2

func MakeClerk(host uint64, cid uint64) *Clerk {
	ck := new(Clerk)
	ck.host = host
	ck.client = MakeRPCClient(host, cid)
	return ck
}

func (ck *Clerk) TryLock(lockname uint64) bool {
	return ck.client.MakeRequest(LOCK_TRYLOCK, grove_common.RPCVals{U64_1: lockname}) != 0
}

//
// ask the lock service to unlock a lock.
// returns true if the lock was previously held,
// false otherwise.
//
func (ck *Clerk) Unlock(lockname uint64) bool {
	return ck.client.MakeRequest(LOCK_UNLOCK, grove_common.RPCVals{U64_1: lockname}) != 0
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
