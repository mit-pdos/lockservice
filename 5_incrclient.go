package lockservice

import (
	"github.com/mit-pdos/lockservice/grove_common"
)

type IncrClerk struct {
	host   string // IncrServer
	client *RPCClient
	cid    uint64
	seq    uint64
}

func MakeIncrClerk(host string, cid uint64) *IncrClerk {
	ck := new(IncrClerk)
	ck.host = host
	ck.client = MakeRPCClient(host, cid)
	return ck
}

func (ck *IncrClerk) Increment(key uint64) {
	ck.client.MakeRequest(IS_INCR, grove_common.RPCVals{U64_1: key})
	return // For goose
}
