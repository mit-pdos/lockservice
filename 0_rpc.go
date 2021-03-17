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

// TODO: rename this stuff to refer to "reply table" or some such
type RpcCoreHandler func(args grove_common.RPCVals) uint64

// maps r -> d^{-1} r d, where d is the decode function and d^{-1} is the encode
// function
func ConjugateRpcFunc(r grove_common.RpcFunc) grove_common.RawRpcFunc {
	return func(rawReq []byte, rawRep *[]byte) bool {
		req := new(grove_common.RPCRequest)
		rep := new(grove_common.RPCReply)
		rpcReqDecode(rawReq, req)
		ret := r(req, rep)
		*rawRep = rpcReplyEncode(rep)
		return ret
	}
}

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

// Common code for RPC clients: tracking of CID and next sequence number.
type RPCClient struct {
	cid uint64
	seq uint64
	rawCl *grove_ffi.RPCClient
}

func MakeRPCClient(host string, cid uint64) *RPCClient {
	return &RPCClient{cid: cid, seq: 1, rawCl:grove_ffi.MakeRPCClient(host)}
}

// 2 refers to the number of u64s in args
func RemoteProcedureCall2(cl *grove_ffi.RPCClient, rpcid uint64, req *grove_common.RPCRequest, reply *grove_common.RPCReply) bool {
	rawReq := rpcReqEncode(req)
	rawRep := make([]byte, 0)
	errb := cl.RemoteProcedureCall(rpcid, &rawReq, &rawRep)
	reply = new(grove_common.RPCReply)
	rpcReplyDecode(rawRep, reply)
	return errb
}

func (cl *RPCClient) MakeRequest(rpcid uint64, args grove_common.RPCVals) uint64 {
	overflow_guard_incr(cl.seq)
	// prepare the arguments.
	req := &grove_common.RPCRequest{Args: args, CID: cl.cid, Seq: cl.seq}
	cl.seq = cl.seq + 1

	rawReq := rpcReqEncode(req)
	rawRep := make([]byte, 0)
	// send an RPC request, wait for the reply.
	var errb = false
	for {
		errb = cl.rawCl.RemoteProcedureCall(rpcid, &rawReq, &rawRep)
		if errb == false {
			break
		}
		continue
	}
	reply := new(grove_common.RPCReply)
	rpcReplyDecode(rawRep, reply)
	return reply.Ret
}

// Common code for RPC servers: handling of stale and redundant requests through
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
