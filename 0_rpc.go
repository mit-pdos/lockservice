package lockservice

import (
	"sync"
)

//
// Common definitions for our RPC layer
//

type RPCArgs struct {
	Arg1 uint64
	Arg2 uint64
}

type RPCRequest struct {
	// Go's net/rpc requires that these field
	// names start with upper case letters!
	CID      uint64
	Seq      uint64
	Args     RPCArgs
}
type RPCReply struct {
	Stale bool
	Ret uint64
}

type RpcFunc func(*RPCRequest, *RPCReply) bool

type RpcCoreHandler func(args RPCArgs) uint64

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

// Emulate an RPC call over a lossy network.
// Returns true iff server reported error or request "timed out".
// For the "real thing", this should instead submit a request via the network.
func RemoteProcedureCall(rpc RpcFunc, req *RPCRequest, reply *RPCReply) bool {
	go func() {
		dummy_reply := new(RPCReply)
		for {
			rpc(req, dummy_reply)
		}
	}()

	if nondet() {
		return rpc(req, reply)
	}
	return true
}

// Common code for RPC clients: tracking of CID and next sequence number.
type RPCClient struct {
	cid     uint64
	seq     uint64
}

func MakeRPCClient(cid uint64) *RPCClient {
	return &RPCClient{cid: cid, seq: 1}
}

func (cl *RPCClient) MakeRequest(rpc RpcFunc, args RPCArgs) uint64 {
	overflow_guard_incr(cl.seq)
	// prepare the arguments.
	var req = &RPCRequest{Args: args, CID: cl.cid, Seq: cl.seq}
	cl.seq = cl.seq + 1

	// send an RPC request, wait for the reply.
	var errb = false
	reply := new(RPCReply)
	for {
		errb = RemoteProcedureCall(rpc, req, reply)
		if errb == false {
			break
		}
		continue
	}
	return reply.Ret
}

// Common code for RPC servers: locking and handling of stale and redundant requests through
// reply "cache".
type RPCServer struct {
	mu *sync.Mutex

	// each CID's last sequence # and the corresponding reply
	lastSeq map[uint64]uint64
	lastReply map[uint64]uint64
}

func MakeRPCServer() *RPCServer {
	sv := new(RPCServer)
	sv.lastSeq = make(map[uint64]uint64)
	sv.lastReply = make(map[uint64]uint64)
	sv.mu = new(sync.Mutex)
	return sv
}

func (sv *RPCServer) HandleRequest(core RpcCoreHandler, req *RPCRequest, reply *RPCReply) bool {
	sv.mu.Lock()

	if CheckReplyCache(sv.lastSeq, sv.lastReply, req.CID, req.Seq, reply) {
		sv.mu.Unlock()
		return false
	}

	reply.Ret = core(req.Args)
	sv.lastReply[req.CID] = reply.Ret
	sv.mu.Unlock()
	return false
}
