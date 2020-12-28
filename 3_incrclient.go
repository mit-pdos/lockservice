package lockservice

type IncrClerk struct {
	primary *IncrServer
	client  *RPCClient
	cid     uint64
	seq     uint64
}

func MakeIncrClerk(primary *IncrServer, cid uint64) *IncrClerk {
	ck := new(IncrClerk)
	ck.primary = primary
	ck.client = MakeRPCClient(cid)
	return ck
}

func (ck *IncrClerk) Increment(key uint64) {
	ck.client.MakeRequest(ck.primary.Increment, RPCVals{U64_1: key})
	return // For goose
}
