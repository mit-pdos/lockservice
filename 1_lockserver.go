package lockservice

import (
	"sync"
)

type LockServer struct {
	mu *sync.Mutex
	// for each lock name, is it locked?
	locks map[uint64]bool

	// each CID's last sequence #
	lastSeq   map[uint64]uint64
	lastReply map[uint64]uint64
}

func (ls *LockServer) tryLock_core(lockname uint64) uint64 {
	locked, _ := ls.locks[lockname]
	if locked {
		return 0
	} else {
		ls.locks[lockname] = true
		return 1
	}
}

func (ls *LockServer) unlock_core(lockname uint64) uint64 {
	locked, _ := ls.locks[lockname]
	if locked {
		ls.locks[lockname] = false
		return 1
	} else {
		return 0
	}
}

func (ls *LockServer) checkReplyCache(CID uint64, Seq uint64, reply *RPCReply) bool {
	last, ok := ls.lastSeq[CID]
	reply.Stale = false
	if ok && Seq <= last {
		if Seq < last {
			reply.Stale = true
			return true
		}
		reply.Ret = ls.lastReply[CID]
		return true
	}
	ls.lastSeq[CID] = Seq
	return false
}

//
// server Lock RPC handler.
// returns true iff error
//
func (ls *LockServer) TryLock(req *TryLockRequest, reply *RPCReply) bool {
	ls.mu.Lock()

	if ls.checkReplyCache(req.CID, req.Seq, reply) {
		ls.mu.Unlock()
		return false
	}
	reply.Ret = ls.tryLock_core(req.Args)

	ls.lastReply[req.CID] = reply.Ret
	ls.mu.Unlock()
	return false
}

//
// server Unlock RPC handler.
// returns true iff error
//
func (ls *LockServer) Unlock(req *UnlockRequest, reply *RPCReply) bool {
	ls.mu.Lock()

	if ls.checkReplyCache(req.CID, req.Seq, reply) {
		ls.mu.Unlock()
		return false
	}

	reply.Ret = ls.unlock_core(req.Args)
	ls.lastReply[req.CID] = reply.Ret
	ls.mu.Unlock()
	return false
}

func MakeLockServer() *LockServer {
	ls := new(LockServer)
	ls.locks = make(map[uint64]bool)

	ls.lastSeq = make(map[uint64]uint64)
	ls.lastReply = make(map[uint64]uint64)
	ls.mu = new(sync.Mutex)
	return ls
}
