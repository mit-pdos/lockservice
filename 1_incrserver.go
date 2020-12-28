package lockservice

import (
	ffi "./grove_ffi"
	"fmt"
	"github.com/tchajed/marshal"
)

type IncrServer struct {
	sv *RPCServer

	// Address of the KVServer; for now a pointer, but in real code would be IP
	// address, etc.
	kvserver *KVServer

	// guarded by sv.mu
	// TODO: Clerks need to be made crash safe, by ensuring that seq's don't get reused
	kck *KVClerk
}

func (is *IncrServer) increment_core_unsafe(seq uint64, args RPCVals) uint64 {
	key := args.U64_1       // A
	var oldv uint64         // A
	oldv = is.kck.Get(key)  // A

	is.kck.Put(key, oldv+1) // B
	return 0
}

// crash-safely increment counter and return the new value
//
// Idea is this:
// In the unsafe version, we ought to have the quadruples
//
// { key [kv]-> a }
//  A
// { key [kv]-> _ \ast v = a }
// { key [kv]-> a }
//
// { key [kv]-> _ \ast v = a \ast durable_oldv = a }
//  B
// { key [kv]-> (a + 1) }
// { key [kv]-> _ \ast durable_oldv = a }
//
// By adding code between A and B that makes durable_oldv = v, we can ensure
// that rerunning the function will result in B starting with the correct
// durable_oldv.
// TODO: test this
// Probably won't try proving this version correct (first).
//
func (is *IncrServer) increment_core(seq uint64, args RPCVals) uint64 {
	key := args.U64_1
	var oldv uint64
	var enc marshal.Enc

	filename := "incr_request_" + fmt.Sprint(seq) + "_oldv"
	if content := ffi.Read(filename); len(content) > 0 {
		oldv = marshal.NewDec(ffi.Read(filename)).GetInt()
		goto about_to_put
	}

	// XXX: This would be annoying to prove correct because the kck.Get() will
	// blindly increment
	// Basically, if we ever crash, we need to give pack the P to our caller;
	// that is, we need the kv ptsto prop. To get this, we would need to
	// "downgrade" the RPCRequestInvariant of the Get() that we're trying to do
	// to have Pre=True and Post=True.
	oldv = is.kck.Get(key)

	enc = marshal.NewEnc(8)
	ffi.Write(filename, enc.Finish())

about_to_put:
	is.kck.Put(key, oldv+1)
	return 0
}

func (is *IncrServer) Increment(req *RPCRequest, reply *RPCReply) bool {
	is.sv.mu.Lock()

	if CheckReplyTable(is.sv.lastSeq, is.sv.lastReply, req.CID, req.Seq, reply) {
	} else {
		reply.Ret = is.increment_core(req.Seq, req.Args)
		is.sv.lastReply[req.CID] = reply.Ret

		// Want this here to write to the reply table
		WriteDurableIncrServer(is)
	}

	is.sv.mu.Unlock()
	return false
}

// For now, there is only one kv server in the whole world
func WriteDurableIncrServer(ks *IncrServer) {
	// TODO: implement persister
	return
}

func ReadDurableIncrServer() *IncrServer {
	// TODO: implement persister
	return nil
}

func MakeIncrServer(kvserver *KVServer) *IncrServer {
	// If we alreay have some saved state, continue from there
	is_old := ReadDurableIncrServer()
	if is_old != nil {
		return is_old
	}

	// Otherwise, we should make a brand new object
	is := new(IncrServer)
	is.sv = MakeRPCServer()
	is.kvserver = kvserver
	return is
}
