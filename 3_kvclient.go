package lockservice

//
// the lockservice Clerk lives in the client
// and maintains a little state.
//
type KVClerk struct {
	primary *KVServer
	cid     uint64
	seq     uint64
}

func MakeKVClerk(primary *KVServer, cid uint64) *KVClerk {
	ck := new(KVClerk)
	ck.primary = primary
	ck.cid = cid
	ck.seq = 1
	return ck
}

func (ck *KVClerk) Put(key uint64, val uint64) {
    overflow_guard_incr(ck.seq)
	// prepare the arguments.
	var args = &PutRequest{Args: PutArgs{Key: key, Value:val}, CID: ck.cid, Seq: ck.seq}
	ck.seq = ck.seq + 1

	// send an RPC request, wait for the reply.
	reply := new(RPCReply)
	for CallPut(ck.primary, args, reply) == true { }
}

func (ck *KVClerk) Get(key uint64) uint64 {
    overflow_guard_incr(ck.seq)
	// prepare the arguments.
	var args = &GetRequest{Args: key, CID: ck.cid, Seq: ck.seq}
	ck.seq = ck.seq + 1

	// send an RPC request, wait for the reply.
	reply := new(RPCReply)
	for CallGet(ck.primary, args, reply) == true { }
	return reply.Ret
}
