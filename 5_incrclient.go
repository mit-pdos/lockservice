package lockservice

import (
	"github.com/mit-pdos/lockservice/grove_common"
)

type IncrClerk struct {
	primary uint64
	client  *RPCClient
	cid     uint64
	seq     uint64
}

const INCR_INCREMENT uint64 = 1

func MakeIncrClerk(primary uint64, cid uint64) *IncrClerk {
	ck := new(IncrClerk)
	ck.primary = primary
	ck.client = MakeRPCClient(cid)
	return ck
}

func (ck *IncrClerk) Increment(key uint64) {
	ck.client.MakeRequest(ck.primary, INCR_INCREMENT, grove_common.RPCVals{U64_1: key})
	return // For goose
}
