package lockservice

type LockServer struct {
	sv *RPCServer
	// for each lock name, is it locked?
	locks map[uint64]bool
}

func (ls *LockServer) tryLock_core(args RPCVals) uint64 {
	lockname := args.U64_1
	locked, _ := ls.locks[lockname]
	if locked {
		return 0
	} else {
		ls.locks[lockname] = true
		return 1
	}
}

func (ls *LockServer) unlock_core(args RPCVals) uint64 {
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
// returns true iff error
//
func (ls *LockServer) TryLock(req *RPCRequest, reply *RPCReply) bool {
	f := func(args RPCVals) uint64 {
		return ls.tryLock_core(args)
	}
	return ls.sv.HandleRequest(f, req, reply)
}

//
// server Unlock RPC handler.
// returns true iff error
//
func (ls *LockServer) Unlock(req *RPCRequest, reply *RPCReply) bool {
	f := func(args RPCVals) uint64 {
		return ls.unlock_core(args)
	}
	return ls.sv.HandleRequest(f, req, reply)
}

func MakeLockServer() *LockServer {
	ls := new(LockServer)
	ls.locks = make(map[uint64]bool)
	ls.sv = MakeRPCServer()
	return ls
}
