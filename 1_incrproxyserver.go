package lockservice

import (
	ffi "./grove_ffi"
	"github.com/tchajed/marshal"
	"fmt"
)

type IncrProxyServer struct {
	sv *RPCServer

	incrserver *IncrServer
	ick        *IncrClerk
}

func (is *IncrProxyServer) proxy_increment_core_unsafe(seq uint64, args RPCVals) uint64 {
	key := args.U64_1
	is.ick.Increment(key)
	return 0
}

// Common code for RPC clients: tracking of CID and next sequence number.
type ShortTermIncrClerk struct {
	cid        uint64
	seq        uint64
	req        RPCRequest
	incrserver *IncrServer
}

func (ck *ShortTermIncrClerk) PrepareRequest(args RPCVals) {
	overflow_guard_incr(ck.seq)
	// prepare the arguments.
	ck.req = RPCRequest{Args: args, CID: ck.cid, Seq: ck.seq}
	ck.seq = ck.seq + 1

}

func (ck *ShortTermIncrClerk) MakePreparedRequest() uint64 {
	// send the already-prepared RPC request, wait for the reply.
	var errb = false
	reply := new(RPCReply)
	for {
		errb = RemoteProcedureCall(ck.incrserver.Increment, &ck.req, reply)
		if errb == false {
			break
		}
		continue
	}
	return reply.Ret
}

func DecodeShortTermIncrClerk(is *IncrServer, content []byte) ShortTermIncrClerk {
	d := marshal.NewDec(content)
	var ck ShortTermIncrClerk
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

func MakeFreshIncrClerk() ShortTermIncrClerk {
	// TODO: get fresh cid
	cid := uint64(0)
	ck := ShortTermIncrClerk{cid:cid}
	return ck
}

func (is *IncrProxyServer) proxy_increment_core(seq uint64, args RPCVals) uint64 {
	filename := "procy_incr_request_" + fmt.Sprint(seq)
	var ck ShortTermIncrClerk

	if content := ffi.Read(filename); len(content) > 0 {
		ck = DecodeShortTermIncrClerk(is.incrserver, content)
	} else {
		ck = MakeFreshIncrClerk()
		ck.PrepareRequest(args)
		content = EncodeShortTermIncrClerk(&ck)
		ffi.Write(filename, content)
	}

	ck.MakePreparedRequest()
	return 0
}
