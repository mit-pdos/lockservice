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

type SrvFunc func(*RPCRequest, *RPCReply) bool

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

// Emulate an RPC call over a lossy network
// Returns true iff server reported error or request "timed out"
func CallRpc(srv SrvFunc, req *RPCRequest, reply *RPCReply) bool {
	go func() {
		dummy_reply := new(RPCReply)
		for {
			srv(req, dummy_reply)
		}
	}()

	if nondet() {
		return srv(req, reply)
	}
	return true
}

type RPCClient struct {
	cid     uint64
	seq     uint64
}

func MakeRPCClient(cid uint64) *RPCClient {
	return &RPCClient{cid: cid, seq: 1}
}

func (cl *RPCClient) MakeRequest(srv SrvFunc, arg1 uint64, arg2 uint64) uint64 {
	overflow_guard_incr(cl.seq)
	// prepare the arguments.
	var args = &RPCRequest{Arg1: arg1, Arg2: arg2, CID: cl.cid, Seq: cl.seq}
	cl.seq = cl.seq + 1

	// send an RPC request, wait for the reply.
	var errb = false
	reply := new(RPCReply)
	for {
		errb = CallRpc(srv, args, reply)
		if errb == false {
			break
		}
		continue
	}
	return reply.Ret
}
