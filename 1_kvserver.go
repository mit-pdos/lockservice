package lockservice

type KVServer struct {
	sv *RPCServer
	// for each lock name, is it locked?
	kvs map[uint64]uint64
}

func (ks *KVServer) put_core(args RPCVals) uint64 {
	ks.kvs[args.U64_1] = args.U64_2
	return 0
}

func (ks *KVServer) get_core(args RPCVals) uint64 {
	return ks.kvs[args.U64_1]
}

// For now, there is only one kv server in the whole world
func WriteDurableKVServer(ks *KVServer) {
	// TODO: implement persister
	return
}

func ReadDurableKVServer() *KVServer {
	// TODO: implement persister
	return nil
}

func (ks *KVServer) Put(req *RPCRequest, reply *RPCReply) bool {
	f := func(args RPCVals) uint64 {
		return ks.put_core(args)
	}
	fdur := func() {
		WriteDurableKVServer(ks)
	}
	r := ks.sv.HandleRequest(f, fdur, req, reply)
	return r
}

func (ks *KVServer) Get(req *RPCRequest, reply *RPCReply) bool {
	f := func(args RPCVals) uint64 {
		return ks.get_core(args)
	}
	fdur := func() {
		WriteDurableKVServer(ks)
	}
	r := ks.sv.HandleRequest(f, fdur, req, reply)
	return r
}

func MakeKVServer() *KVServer {
	// If we alreay have some saved state, continue from there
	ks_old := ReadDurableKVServer()
	if ks_old != nil {
		return ks_old
	}

	// Otherwise, we should make a brand new object
	ks := new(KVServer)
	ks.kvs = make(map[uint64]uint64)
	ks.sv = MakeRPCServer()
	return ks
}
