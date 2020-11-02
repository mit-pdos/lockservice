package lockservice

//
// Common definitions for our RPC layer
//

type RPCRequest struct {
	// Go's net/rpc requires that these field
	// names start with upper case letters!
	CID      uint64
	Seq      uint64
	Arg1 uint64
	Arg2 uint64
}
type RPCReply struct {
	Stale bool
	Ret uint64
}

func CheckReplyCache(
	lastSeq map[uint64]uint64,
	lastReply map[uint64]uint64,
	CID uint64,
	Seq uint64,
	reply *RPCReply,
) bool {
	last, ok := lastSeq[CID]
	reply.Stale = false
	if ok && Seq <= last {
		if Seq < last {
			reply.Stale = true
			return true
		}
		reply.Ret = lastReply[CID]
		return true
	}
	lastSeq[CID] = Seq
	return false
}
