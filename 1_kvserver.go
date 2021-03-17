package lockservice

// non-crash safe kv service

import (
	"github.com/mit-pdos/lockservice/grove_common"
	"github.com/mit-pdos/lockservice/grove_ffi"
	"sync"
)

const KV_PUT uint64 = 1
const KV_GET uint64 = 2

type KVServer struct {
	mu *sync.Mutex
	sv *RPCServer
	// for each lock name, is it locked?
	kvs map[uint64]uint64
}

func (ks *KVServer) put_core(args grove_common.RPCVals) uint64 {
	ks.kvs[args.U64_1] = args.U64_2
	return 0
}

func (ks *KVServer) get_core(args grove_common.RPCVals) uint64 {
	return ks.kvs[args.U64_1]
}

func (ks *KVServer) Put(req *grove_common.RPCRequest, reply *grove_common.RPCReply) {
	ks.mu.Lock()
	ks.sv.HandleRequest(ks.put_core, req, reply)
	ks.mu.Unlock()
	return
}

func (ks *KVServer) Get(req *grove_common.RPCRequest, reply *grove_common.RPCReply) {
	ks.mu.Lock()
	ks.sv.HandleRequest(ks.get_core, req, reply)
	ks.mu.Unlock()
	return
}

func MakeKVServer() *KVServer {
	ks := new(KVServer)
	ks.mu = new(sync.Mutex)
	ks.kvs = make(map[uint64]uint64)
	ks.sv = MakeRPCServer()
	return ks
}

func (ks *KVServer) Start() {
	handlers := make(map[uint64]grove_common.RawRpcFunc)
	handlers[KV_PUT] = ConjugateRpcFunc(ks.Put)
	handlers[KV_GET] = ConjugateRpcFunc(ks.Get)
	grove_ffi.StartRPCServer(handlers)
}
