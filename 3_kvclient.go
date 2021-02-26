package lockservice

import (
	"github.com/mit-pdos/lockservice/grove_common"
)

//
// the lockservice Clerk lives in the client
// and maintains a little state.
//
type KVClerk struct {
	primary uint64
	client *RPCClient
	cid     uint64
	seq     uint64
}

func MakeKVClerk(primary uint64, cid uint64) *KVClerk {
	ck := new(KVClerk)
	ck.primary = primary
	ck.client = MakeRPCClient(cid)
	return ck
}

func (ck *KVClerk) Put(key uint64, val uint64) {
	ck.client.MakeRequest(ck.primary, KV_PUT, grove_common.RPCVals{U64_1: key, U64_2: val})
	return // For goose
}

func (ck *KVClerk) Get(key uint64) uint64 {
	return ck.client.MakeRequest(ck.primary, KV_GET, grove_common.RPCVals{U64_1:key})
}
