package lockservice

import (
	"github.com/mit-pdos/lockservice/grove_common"
	"github.com/mit-pdos/lockservice/grove_ffi"
	"github.com/tchajed/goose/machine"
	"github.com/tchajed/marshal"
)

//
// Common definitions for our RPC layer
//

type RpcCoreHandler func(args grove_common.RPCVals) uint64

func CheckReplyTable(
	lastSeq map[uint64]uint64,
	lastReply map[uint64]uint64,
	CID uint64,
	Seq uint64,
	reply *grove_common.RPCReply,
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

func rpcReqEncode(req *grove_common.RPCRequest) []byte {
	e := marshal.NewEnc(4 * 8)
	e.PutInt(req.CID)
	e.PutInt(req.Seq)
	e.PutInt(req.Args.U64_1)
	e.PutInt(req.Args.U64_2)
	res := e.Finish()
	machine.Linearize()
	return res
}

func rpcReqDecode(data []byte, req *grove_common.RPCRequest) {
	d := marshal.NewDec(data)
	req.CID = d.GetInt()
	req.Seq = d.GetInt()
	req.Args.U64_1 = d.GetInt()
	req.Args.U64_2 = d.GetInt()
	machine.Linearize()
}

func rpcReplyEncode(reply *grove_common.RPCReply) []byte {
	e := marshal.NewEnc(2 * 8)
	e.PutBool(reply.Stale)
	e.PutInt(reply.Ret)
	res := e.Finish()
	machine.Linearize()
	return res
}

func rpcReplyDecode(data []byte, reply *grove_common.RPCReply) {
	d := marshal.NewDec(data)
	reply.Stale = d.GetBool()
	reply.Ret = d.GetInt()
	machine.Linearize()
}

// Emulate an RPC call over a lossy network.
// Returns true iff server reported error or request "timed out".
// For the "real thing", this should instead submit a request via the network.
func RemoteProcedureCall(host uint64, rpcid uint64, req *grove_common.RPCRequest, reply *grove_common.RPCReply) bool {
	reqdata := rpcReqEncode(req)

	go func() {
		dummy_reply := new(grove_common.RPCReply)
		for {
			rpc := grove_ffi.GetServer(host, rpcid)
			decodedReq := new(grove_common.RPCRequest)
			rpcReqDecode(reqdata, decodedReq)
			rpc(decodedReq, dummy_reply)
		}
	}()

	if nondet() {
		rpc := grove_ffi.GetServer(host, rpcid)
		decodedReq := new(grove_common.RPCRequest)
		rpcReqDecode(reqdata, decodedReq)

		serverReply := new(grove_common.RPCReply)
		ok := rpc(req, serverReply)

		replydata := rpcReplyEncode(serverReply)
		rpcReplyDecode(replydata, reply)
		return ok
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

func (cl *RPCClient) MakeRequest(host uint64, rpcid uint64, args grove_common.RPCVals) uint64 {
	overflow_guard_incr(cl.seq)
	// prepare the arguments.
	req := &grove_common.RPCRequest{Args: args, CID: cl.cid, Seq: cl.seq}
	cl.seq = cl.seq + 1

	// send an RPC request, wait for the reply.
	var errb = false
	reply := new(grove_common.RPCReply)
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
	// each CID's last sequence # and the corresponding reply
	lastSeq   map[uint64]uint64
	lastReply map[uint64]uint64
}

func MakeRPCServer() *RPCServer {
	sv := new(RPCServer)
	sv.lastSeq = make(map[uint64]uint64)
	sv.lastReply = make(map[uint64]uint64)
	return sv
}

func (sv *RPCServer) HandleRequest(core RpcCoreHandler, req *grove_common.RPCRequest, reply *grove_common.RPCReply) bool {
	if CheckReplyTable(sv.lastSeq, sv.lastReply, req.CID, req.Seq, reply) {
	} else {
		reply.Ret = core(req.Args)
		sv.lastReply[req.CID] = reply.Ret
	}

	return false
}
