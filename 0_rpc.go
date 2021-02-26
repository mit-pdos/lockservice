package lockservice

import (
	"sync"
)

//
// Common definitions for our RPC layer
//

type RPCVals struct {
	U64_1 uint64
	U64_2 uint64
}

type RPCRequest struct {
	// Go's net/rpc requires that these field
	// names start with upper case letters!
	CID  uint64
	Seq  uint64
	Args RPCVals
}
type RPCReply struct {
	Stale bool
	Ret   uint64
}

type RpcFunc func(*RPCRequest, *RPCReply) bool

type RpcCoreHandler func(args RPCVals) uint64

type RpcCorePersister func()

func CheckReplyTable(
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

// rpcHandlers describes the RpcFunc handler that will run
// when an RPC is issued to some machine (first key) to invoke
// a specific RPC number (second key).
var rpcHandlers map[uint64]map[uint64]RpcFunc
var rpcNextHost uint64
var rpcHandlersLock sync.Mutex

func allocServer(handlers map[uint64]RpcFunc) uint64 {
	rpcHandlersLock.Lock()

	id := rpcNextHost
	rpcNextHost = rpcNextHost + 1

	if rpcHandlers == nil {
		rpcHandlers = make(map[uint64]map[uint64]RpcFunc)
	}

	rpcHandlers[id] = handlers
	rpcHandlersLock.Unlock()
	return id
}

// Emulate an RPC call over a lossy network.
// Returns true iff server reported error or request "timed out".
// For the "real thing", this should instead submit a request via the network.
func RemoteProcedureCall(host uint64, rpcid uint64, req *RPCRequest, reply *RPCReply) bool {
	go func() {
		dummy_reply := new(RPCReply)
		for {
			rpcHandlersLock.Lock()
			rpc := rpcHandlers[host][rpcid]
			rpcHandlersLock.Unlock()
			rpc(req, dummy_reply)
		}
	}()

	if nondet() {
		rpcHandlersLock.Lock()
		rpc := rpcHandlers[host][rpcid]
		rpcHandlersLock.Unlock()
		return rpc(req, reply)
	}
	return true
}

// Common code for RPC clients: tracking of CID and next sequence number.
type RPCClient struct {
	cid uint64
	seq uint64
}

func MakeRPCClient(cid uint64) *RPCClient {
	return &RPCClient{cid: cid, seq: 1}
}

func (cl *RPCClient) MakeRequest(host uint64, rpcid uint64, args RPCVals) uint64 {
	overflow_guard_incr(cl.seq)
	// prepare the arguments.
	req := &RPCRequest{Args: args, CID: cl.cid, Seq: cl.seq}
	cl.seq = cl.seq + 1

	// send an RPC request, wait for the reply.
	var errb = false
	reply := new(RPCReply)
	for {
		errb = RemoteProcedureCall(host, rpcid, req, reply)
		if errb == false {
			break
		}
		continue
	}
	return reply.Ret
}

// Common code for RPC servers: locking and handling of stale and redundant requests through
// the reply table.
type RPCServer struct {
	mu *sync.Mutex

	// each CID's last sequence # and the corresponding reply
	lastSeq   map[uint64]uint64
	lastReply map[uint64]uint64
}

func MakeRPCServer() *RPCServer {
	sv := new(RPCServer)
	sv.lastSeq = make(map[uint64]uint64)
	sv.lastReply = make(map[uint64]uint64)
	sv.mu = new(sync.Mutex)
	return sv
}

func (sv *RPCServer) HandleRequest(core RpcCoreHandler, makeDurable RpcCorePersister, req *RPCRequest, reply *RPCReply) bool {
	sv.mu.Lock()

	if CheckReplyTable(sv.lastSeq, sv.lastReply, req.CID, req.Seq, reply) {
	} else {
		reply.Ret = core(req.Args)
		sv.lastReply[req.CID] = reply.Ret

		makeDurable()
	}

	sv.mu.Unlock()
	return false
}
