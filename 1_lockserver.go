package lockservice

// non-crash safe lockservice

import (
	"github.com/mit-pdos/lockservice/grove_common"
	"github.com/mit-pdos/lockservice/grove_ffi"
	"sync"
)

type LockServer struct {
	mu *sync.Mutex
	sv *RPCServer
	// for each lock name, is it locked?
	locks map[uint64]bool
}

func (ls *LockServer) tryLock_core(args grove_common.RPCVals) uint64 {
	lockname := args.U64_1
	locked, _ := ls.locks[lockname]
	if locked {
		return 0
	} else {
		ls.locks[lockname] = true
		return 1
	}
}

func (ls *LockServer) unlock_core(args grove_common.RPCVals) uint64 {
	lockname := args.U64_1
	locked, _ := ls.locks[lockname]
	if locked {
		ls.locks[lockname] = false
		return 1
	} else {
		return 0
	}
}

//
// server Lock RPC handler.
//
func (ls *LockServer) TryLock(req *grove_common.RPCRequest, reply *grove_common.RPCReply) {
	ls.mu.Lock()
	ls.sv.HandleRequest(ls.tryLock_core, req, reply)
	ls.mu.Unlock()
	return
}

//
// server Unlock RPC handler.
//
func (ls *LockServer) Unlock(req *grove_common.RPCRequest, reply *grove_common.RPCReply) {
	ls.mu.Lock()
	ls.sv.HandleRequest(ls.unlock_core, req, reply)
	ls.mu.Unlock()
	return
}

func MakeLockServer() *LockServer {
	ls := new(LockServer)
	ls.mu = new(sync.Mutex)
	ls.locks = make(map[uint64]bool)
	ls.sv = MakeRPCServer()
	return ls
}

func (ls *LockServer) Start() {
	handlers := make(map[uint64]grove_common.RawRpcFunc)
	handlers[LOCK_TRYLOCK] = ConjugateRpcFunc(ls.TryLock)
	handlers[LOCK_UNLOCK] = ConjugateRpcFunc(ls.Unlock)
	grove_ffi.StartRPCServer(handlers)
}
