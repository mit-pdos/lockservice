package lockservice

type KVServer struct {
	sv *RPCServer
	// for each lock name, is it locked?
	kvs map[uint64]uint64
}

func (ks *KVServer) put_core(args RPCArgs) uint64 {
	ks.kvs[args.Arg1] = args.Arg2
	return 0
}

func (ks *KVServer) get_core(args RPCArgs) uint64 {
	return ks.kvs[args.Arg1]
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
