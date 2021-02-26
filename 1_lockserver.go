package lockservice

import (
	"github.com/mit-pdos/lockservice/grove_common"
)

type LockServer struct {
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

// For now, there is only one lock server in the whole world
func WriteDurableLockServer(ks *LockServer) {
	// TODO: implement persister
	return
}

func ReadDurableLockServer() *LockServer {
	// TODO: implement persister
	return nil
}

//
// server Lock RPC handler.
// returns true iff error
//
func (ls *LockServer) TryLock(req *grove_common.RPCRequest, reply *grove_common.RPCReply) bool {
	f := func(args grove_common.RPCVals) uint64 {
		return ls.tryLock_core(args)
	}
	fdur := func() {
		WriteDurableLockServer(ls)
	}
	r := ls.sv.HandleRequest(f, fdur, req, reply)
	WriteDurableLockServer(ls)
	return r
}

//
// server Unlock RPC handler.
// returns true iff error
//
func (ls *LockServer) Unlock(req *grove_common.RPCRequest, reply *grove_common.RPCReply) bool {
	f := func(args grove_common.RPCVals) uint64 {
		return ls.unlock_core(args)
	}
	fdur := func() {
		WriteDurableLockServer(ls)
	}
	r := ls.sv.HandleRequest(f, fdur, req, reply)
	WriteDurableLockServer(ls)
	return r
}

func MakeLockServer() *LockServer {
	ls_old := ReadDurableLockServer()
	if ls_old != nil {
		return ls_old
	}

	ls := new(LockServer)
	ls.locks = make(map[uint64]bool)
	ls.sv = MakeRPCServer()
	return ls
}
