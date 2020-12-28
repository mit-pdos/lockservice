package lockservice

import (
	ffi "./grove_ffi"
)

type IncrServer struct {
	sv *RPCServer

	// Address of the KVServer; for now a pointer, but in real code would be IP
	// address, etc.
	kvserver *KVServer
}

// crash-safely increment counter and return the new value
func (is *IncrServer) increment_core(args RPCVals) uint64 {
	// TODO: impl
	ffi.Write("", nil)
	return 0
}


func (is *IncrServer) Increment(req *RPCRequest, reply *RPCReply) bool {
	f := func(args RPCVals) uint64 {
		return is.increment_core(args)
	}
	fdur := func() {
		WriteDurableIncrServer(is)
	}
	r := is.sv.HandleRequest(f, fdur, req, reply)
	return r
}


// For now, there is only one kv server in the whole world
func WriteDurableIncrServer(ks *IncrServer) {
	// TODO: implement persister
	return
}

func ReadDurableIncrServer() *IncrServer {
	// TODO: implement persister
	return nil
}

func MakeIncrServer(kvserver *KVServer) *IncrServer {
	// If we alreay have some saved state, continue from there
	is_old := ReadDurableIncrServer()
	if is_old != nil {
		return is_old
	}

	// Otherwise, we should make a brand new object
	is := new(IncrServer)
	is.sv = MakeRPCServer()
	is.kvserver = kvserver
	return is
}
