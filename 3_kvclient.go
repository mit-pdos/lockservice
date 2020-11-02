package lockservice

//
// the lockservice Clerk lives in the client
// and maintains a little state.
//
type KVClerk struct {
	primary *KVServer
	client *RPCClient
	cid     uint64
	seq     uint64
}

func MakeKVClerk(primary *KVServer, cid uint64) *KVClerk {
	ck := new(KVClerk)
	ck.primary = primary
	ck.client = MakeRPCClient(cid)
	return ck
}

func (ck *KVClerk) Put(key uint64, val uint64) {
	ck.client.MakeRequest(ck.primary.Put, RPCArgs{Arg1:key, Arg2:val})
}

func (ck *KVClerk) Get(key uint64) uint64 {
	return ck.client.MakeRequest(ck.primary.Get, RPCArgs{Arg1:key})
}
