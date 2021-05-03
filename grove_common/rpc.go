package grove_common

type RawRPCRequest struct {
	RpcId uint64
	Data  []byte
}

type RawRPCReply struct {
	Data []byte
}

type RawRpcFunc func([]byte, *[]byte)

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

type RpcFunc func(*RPCRequest, *RPCReply)
