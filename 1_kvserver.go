package lockservice

type KVServer struct {
	sv *RPCServer
	// for each lock name, is it locked?
	kvs map[uint64]uint64
}

func (ks *KVServer) put_core(key uint64, value uint64) uint64 {
	ks.kvs[key] = value
	return 0
}

func (ks *KVServer) get_core(key uint64, dummy uint64) uint64 {
	return ks.kvs[key]
}



func (ks *KVServer) Put(req *RPCRequest, reply *RPCReply) bool {
	return ks.sv.HandleRequest(ks.put_core, req, reply)
}

func (ks *KVServer) Get(req *RPCRequest, reply *RPCReply) bool {
	return ks.sv.HandleRequest(ks.get_core, req, reply)
}

func MakeKVServer() *KVServer {
	ks := new(KVServer)
	ks.kvs = make(map[uint64]uint64)
	ks.sv = MakeRPCServer()
	return ks
}
