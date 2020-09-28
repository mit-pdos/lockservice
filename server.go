package lockservice

import (
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"
)

type LockServer struct {
	mu	  *sync.Mutex
	l	  net.Listener

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
	if ok && args.Seq <= last {
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
	if ok && args.Seq <= last {
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

func (ls *LockServer) kill() {
	ls.l.Close()
}

func StartServer(primary string) (*LockServer) {
	ls := new(LockServer)
	ls.locks = make(map[uint64]bool)

	ls.lastSeq = make(map[uint64]uint64)
	ls.lastReply = make(map[uint64]bool)

	me := primary

	// tell net/rpc about our RPC server and handlers.
	rpcs := rpc.NewServer()
	rpcs.Register(ls)

	// prepare to receive connections from clients.
	// change "unix" to "tcp" to use over a network.
	os.Remove(me) // only needed for "unix"
	l, e := net.Listen("unix", me)
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	ls.l = l

	// please don't change any of the following code,
	// or do anything to subvert it.

	// create a thread to accept RPC connections from clients.
	go func() {
		for {
			conn, err := ls.l.Accept()
			if err == nil {
				go rpcs.ServeConn(conn)
			} else if err == nil {
				conn.Close()
			}
		}
	}()

	return ls
}
