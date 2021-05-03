package lockservice

import (
	"github.com/mit-pdos/lockservice/grove_common"
	"github.com/mit-pdos/lockservice/grove_ffi"
	"github.com/tchajed/marshal"
)

//
// Common definitions for our RPC layer
//

// TODO: rename this stuff to refer to "reply table" or some such
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
	return res
}

func rpcReqDecode(data []byte, req *grove_common.RPCRequest) {
	d := marshal.NewDec(data)
	req.CID = d.GetInt()
	req.Seq = d.GetInt()
	req.Args.U64_1 = d.GetInt()
	req.Args.U64_2 = d.GetInt()
}

func rpcReplyEncode(reply *grove_common.RPCReply) []byte {
	e := marshal.NewEnc(2 * 8)
	e.PutBool(reply.Stale)
	e.PutInt(reply.Ret)
	res := e.Finish()
	return res
}

func rpcReplyDecode(data []byte, reply *grove_common.RPCReply) {
	d := marshal.NewDec(data)
	reply.Stale = d.GetBool()
	reply.Ret = d.GetInt()
}

// Essentially maps r -> e r e^{-1}, where e is the encode function and e^{-1} is decode
// function
func ConjugateRpcFunc(r grove_common.RpcFunc) grove_common.RawRpcFunc {
	return func(rawReq []byte, rawRep *[]byte) {
		req := new(grove_common.RPCRequest)
		rep := new(grove_common.RPCReply)
		rpcReqDecode(rawReq, req)
		r(req, rep)
		*rawRep = rpcReplyEncode(rep)
		return
	}
}

// Common code for RPC clients: tracking of CID and next sequence number.
type RPCClient struct {
	cid   uint64
	seq   uint64
	rawCl *grove_ffi.RPCClient
}

func MakeRPCClient(host uint64, cid uint64) *RPCClient {
	return &RPCClient{cid: cid, seq: 1, rawCl: grove_ffi.MakeRPCClient(host)}
}

// 2 refers to the number of u64s in args
func RemoteProcedureCall2(cl *grove_ffi.RPCClient, rpcid uint64, req *grove_common.RPCRequest, rep *grove_common.RPCReply) bool {
	rawReq := rpcReqEncode(req)
	rawRep := new([]byte)
	errb := cl.Call(rpcid, rawReq, rawRep)
	if errb == false {
		rpcReplyDecode(*rawRep, rep)
	}
	return errb
}

func (cl *RPCClient) MakeRequest(rpcid uint64, args grove_common.RPCVals) uint64 {
	overflow_guard_incr(cl.seq)
	// prepare the arguments.
	req := &grove_common.RPCRequest{Args: args, CID: cl.cid, Seq: cl.seq}
	reply := new(grove_common.RPCReply)
	cl.seq = cl.seq + 1

	var errb = false
	for {
		errb = RemoteProcedureCall2(cl.rawCl, rpcid, req, reply)
		if errb == false {
			break
		}
		continue
	}
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

func (sv *RPCServer) HandleRequest(core RpcCoreHandler, req *grove_common.RPCRequest, reply *grove_common.RPCReply) {
	if CheckReplyTable(sv.lastSeq, sv.lastReply, req.CID, req.Seq, reply) {
	} else {
		reply.Ret = core(req.Args)
		sv.lastReply[req.CID] = reply.Ret
	}

	return
}
