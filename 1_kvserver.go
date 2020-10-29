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

func (ks *KVServer) put_core(kv PutArgs) uint64 {
	ks.kvs[kv.Key] = kv.Value
	return 0
}

func (ks *KVServer) get_core(key uint64) uint64 {
	return ks.kvs[key]
}

func (ks *KVServer) checkReplyCache(CID uint64, Seq uint64, reply *RPCReply) bool {
	last, ok := ks.lastSeq[CID]
	reply.Stale = false
	if ok && Seq <= last {
		if Seq < last {
			reply.Stale = true
			return true
		}
		reply.Ret = ks.lastReply[CID]
		return true
	}
	ks.lastSeq[CID] = Seq
	return false
}

func (ks *KVServer) Put(req *PutRequest, reply *RPCReply) bool {
	ks.mu.Lock()

	if ks.checkReplyCache(req.CID, req.Seq, reply) {
		ks.mu.Unlock()
		return false
	}
	reply.Ret = ks.put_core(req.Args)

	ks.lastReply[req.CID] = reply.Ret
	ks.mu.Unlock()
	return false
}

func (ks *KVServer) Get(req *GetRequest, reply *RPCReply) bool {
	ks.mu.Lock()

	if ks.checkReplyCache(req.CID, req.Seq, reply) {
		ks.mu.Unlock()
		return false
	}

	reply.Ret = ks.get_core(req.Args)
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
