package lockservice

import (
	"github.com/mit-pdos/lockservice/grove_common"
)

type KVServer struct {
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



func (ks *KVServer) Put(req *grove_common.RPCRequest, reply *grove_common.RPCReply) bool {
	return ks.sv.HandleRequest(func(args grove_common.RPCVals) uint64 {
		return ks.put_core(args)
	}, req, reply)
}

func (ks *KVServer) Get(req *grove_common.RPCRequest, reply *grove_common.RPCReply) bool {
	return ks.sv.HandleRequest(func(args grove_common.RPCVals) uint64 {
		return ks.get_core(args)
	}, req, reply)
}

func MakeKVServer() *KVServer {
	ks := new(KVServer)
	ks.kvs = make(map[uint64]uint64)
	ks.sv = MakeRPCServer()
	return ks
}
