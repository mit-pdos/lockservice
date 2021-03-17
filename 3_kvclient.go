package lockservice

// non crash-safe clerk for kv service

import (
	"github.com/mit-pdos/lockservice/grove_common"
)

type KVClerk struct {
	host string
	client  *RPCClient
	cid     uint64
	seq     uint64
}

func MakeKVClerk(host string, cid uint64) *KVClerk {
	ck := new(KVClerk)
	ck.host = host
	ck.client = MakeRPCClient(host, cid)
	ck.cid = cid
	ck.seq = 1
	return ck
}

func (ck *KVClerk) Put(key uint64, val uint64) {
	ck.client.MakeRequest(KV_PUT, grove_common.RPCVals{U64_1: key, U64_2: val})
	return // For goose
}

func (ck *KVClerk) Get(key uint64) uint64 {
	return ck.client.MakeRequest(KV_GET, grove_common.RPCVals{U64_1: key})
}
