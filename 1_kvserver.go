package lockservice

import (
	"sync"
)

type KVServer struct {
	mu *sync.Mutex
	// for each lock name, is it locked?
	kvs map[uint64]uint64

	// each CID's last sequence #
	lastSeq   map[uint64]uint64
	lastReply map[uint64]uint64
}

func (ks *KVServer) put_core(key uint64, value uint64) uint64 {
	ks.kvs[key] = value
	return 0
}

func (ks *KVServer) get_core(key uint64) uint64 {
	return ks.kvs[key]
}



func (ks *KVServer) Put(req *RPCRequest, reply *RPCReply) bool {
	ks.mu.Lock()

	if CheckReplyCache(ks.lastSeq, ks.lastReply, req.CID, req.Seq, reply) {
		ks.mu.Unlock()
		return false
	}
	reply.Ret = ks.put_core(req.Arg1, req.Arg2)

	ks.lastReply[req.CID] = reply.Ret
	ks.mu.Unlock()
	return false
}

func (ks *KVServer) Get(req *RPCRequest, reply *RPCReply) bool {
	ks.mu.Lock()

	if CheckReplyCache(ks.lastSeq, ks.lastReply, req.CID, req.Seq, reply) {
		ks.mu.Unlock()
		return false
	}

	reply.Ret = ks.get_core(req.Arg1)
	ks.lastReply[req.CID] = reply.Ret
	ks.mu.Unlock()
	return false
}

func MakeKVServer() *KVServer {
	ks := new(KVServer)
	ks.kvs = make(map[uint64]uint64)

	ks.lastSeq = make(map[uint64]uint64)
	ks.lastReply = make(map[uint64]uint64)
	ks.mu = new(sync.Mutex)
	return ks
}
