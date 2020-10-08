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
	lastReply map[uint64]bool
}

//
// server Lock RPC handler.
// returns true iff error
//
func (ls *LockServer) TryLock(args *LockArgs, reply *LockReply) bool {
	ls.mu.Lock()

	// Check if seqno has been seen, and reply from the cache if so
	last, ok := ls.lastSeq[args.CID]
	reply.Stale = false
	if ok && args.Seq <= last {
		if args.Seq < last {
			reply.Stale = true
			ls.mu.Unlock()
			return false
		}
		reply.OK = ls.lastReply[args.CID]
		ls.mu.Unlock()
		return false
	}
	ls.lastSeq[args.CID] = args.Seq

	locked, _ := ls.locks[args.Lockname]

	if locked {
		reply.OK = false
	} else {
		reply.OK = true
		ls.locks[args.Lockname] = true
	}
	ls.lastReply[args.CID] = reply.OK
	ls.mu.Unlock()
	return false
}

//
// server Unlock RPC handler.
// returns true iff error
//
func (ls *LockServer) Unlock(args *UnlockArgs, reply *UnlockReply) bool {
	ls.mu.Lock()

	last, ok := ls.lastSeq[args.CID]
	reply.Stale = false
	if ok && args.Seq <= last {
		if args.Seq < last {
			reply.Stale = true
			ls.mu.Unlock()
			return false
		}
		reply.OK = ls.lastReply[args.CID]
		ls.mu.Unlock()
		return false
	}
	ls.lastSeq[args.CID] = args.Seq

	locked, _ := ls.locks[args.Lockname]

	if locked {
		ls.locks[args.Lockname] = false
		reply.OK = true
	} else {
		reply.OK = false
	}
	ls.lastReply[args.CID] = reply.OK
	ls.mu.Unlock()
	return false
}

func MakeServer() *LockServer {
	ls := new(LockServer)
	ls.locks = make(map[uint64]bool)

	ls.lastSeq = make(map[uint64]uint64)
	ls.lastReply = make(map[uint64]bool)
	ls.mu = new(sync.Mutex)
	return ls
}
