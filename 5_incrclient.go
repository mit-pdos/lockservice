package lockservice

import (
	"github.com/mit-pdos/lockservice/grove_common"
)

type IncrClerk struct {
	srv uint64 // IncrServer
	client  *RPCClient
	cid     uint64
	seq     uint64
}

func MakeIncrClerk(srv uint64, cid uint64) *IncrClerk {
	ck := new(IncrClerk)
	ck.srv = srv
	ck.client = MakeRPCClient(cid)
	return ck
}

func (ck *IncrClerk) Increment(key uint64) {
	ck.client.MakeRequest(ck.srv, IS_INCR, grove_common.RPCVals{U64_1: key})
	return // For goose
}
