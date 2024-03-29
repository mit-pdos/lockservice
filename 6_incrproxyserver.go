package lockservice

import (
	"github.com/mit-pdos/lockservice/grove_common"
	"github.com/mit-pdos/lockservice/grove_ffi"
	"github.com/tchajed/marshal"
)

type IncrProxyServer struct {
	sv *RPCServer

	incrserver HostName
	ick        *IncrClerk
	lastCID    uint64
}

func (is *IncrProxyServer) proxy_increment_core_unsafe(seq uint64, args grove_common.RPCVals) uint64 {
	key := args.U64_1
	is.ick.Increment(key)
	return 0
}

// Common code for RPC clients: tracking of CID and next sequence number.
type ShortTermIncrClerk struct {
	cid        uint64
	seq        uint64
	req        grove_common.RPCRequest
	incrserver HostName
}

func (ck *ShortTermIncrClerk) PrepareRequest(args grove_common.RPCVals) {
	overflow_guard_incr(ck.seq)
	// prepare the arguments.
	ck.req = grove_common.RPCRequest{Args: args, CID: ck.cid, Seq: ck.seq}
	ck.seq = ck.seq + 1

}

func (ck *ShortTermIncrClerk) MakePreparedRequest() uint64 {
	// send the already-prepared RPC request, wait for the reply.
	var errb = false
	reply := new(grove_common.RPCReply)
	cl := grove_ffi.MakeRPCClient(ck.incrserver)
	for {
		errb = RemoteProcedureCall2(cl, IS_INCR, &ck.req, reply)
		if errb == false {
			break
		}
		continue
	}
	return reply.Ret
}

func DecodeShortTermIncrClerk(is HostName, content []byte) *ShortTermIncrClerk {
	d := marshal.NewDec(content)
	ck := new(ShortTermIncrClerk)
	ck.incrserver = is
	ck.cid = d.GetInt()
	ck.seq = d.GetInt()
	ck.req.CID = ck.cid
	ck.req.Seq = ck.seq - 1
	ck.req.Args.U64_1 = d.GetInt()
	ck.req.Args.U64_2 = d.GetInt()
	return ck
}

func EncodeShortTermIncrClerk(ck *ShortTermIncrClerk) []byte {
	e := marshal.NewEnc(32) // 4 uint64s
	e.PutInt(ck.cid)
	e.PutInt(ck.seq)
	// e.PutInt(ck.req.CID) // this is redundant;
	// e.PutInt(ck.req.Seq)
	e.PutInt(ck.req.Args.U64_1)
	e.PutInt(ck.req.Args.U64_2)
	return e.Finish()
}

func (is *IncrProxyServer) MakeFreshIncrClerk() *ShortTermIncrClerk {
	cid := uint64(is.lastCID)
	overflow_guard_incr(is.lastCID)
	is.lastCID = is.lastCID + 1

	// Make sure that CIDs don't get reused;
	// IncrProxyServer owns all RPCClient_own for any cid >= lastCID
	e := marshal.NewEnc(8)
	e.PutInt(is.lastCID)
	grove_ffi.Write("lastCID", e.Finish())

	ck_ptr := new(ShortTermIncrClerk)
	ck_ptr.cid = cid
	ck_ptr.seq = 1
	ck_ptr.incrserver = is.incrserver

	return ck_ptr
}

func (is *IncrProxyServer) proxy_increment_core(seq uint64, args grove_common.RPCVals) uint64 {
	filename := "procy_incr_request_" + grove_ffi.U64ToString(seq)
	var ck *ShortTermIncrClerk

	content := grove_ffi.Read(filename)
	if len(content) > 0 {
		ck = DecodeShortTermIncrClerk(is.incrserver, content)
	} else {
		ck = is.MakeFreshIncrClerk()
		ck.PrepareRequest(args)
		content := EncodeShortTermIncrClerk(ck) // this shadows the other content variable
		grove_ffi.Write(filename, content)
	}

	ck.MakePreparedRequest()
	return 0
}
