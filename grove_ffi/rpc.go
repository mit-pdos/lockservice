package grove_ffi

import (
	"fmt"
	"net"
	"net/rpc"
	"log"
	"net/http"
	"github.com/mit-pdos/lockservice/grove_common"
)

var port uint64

// this is NOT exposed in the FFI. This just allows us to have StartServer()
// serve at different ports on the same host without exposing port numbers to
// proof-land
func SetPort(p uint64) {
	port = p
}

// shim so we can use net/rpc
// net/rpc does the work of managing the network connections and matching up
// responses to their requests, which we'll eventually want to do ourselves.
type RPCHandler struct {
	// rpcHandlers describes the RawRpcFunc handler that will run
	// when an RPC with a specific RPC number (the key) is invoked.
	rpcHandlers map[uint64]grove_common.RawRpcFunc
}

func (s *RPCHandler) Handle(req *grove_common.RawRPCRequest, rep *[]byte) error {
	if s.rpcHandlers[req.RpcId](req.Data, rep) { // check for error
		panic("Error in RPC handler")
	}
	return nil
}

// starts running an RPC server with the given functions at the corresponding
// endpoints; never returns
func StartRPCServer(handlers map[uint64]grove_common.RawRpcFunc) {
	s := &RPCHandler{}

	if handlers == nil {
		s.rpcHandlers = make(map[uint64]grove_common.RawRpcFunc)
	} else {
		s.rpcHandlers = handlers
	}

	rpc.Register(s)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if e != nil {
		panic(e)
	}
	func() {
		log.Fatal(http.Serve(l, nil))
	}()
}

type RPCClient struct {
	rpc.Client
}

func MakeRPCClient(host string) *RPCClient {
	cl, err := rpc.DialHTTP("tcp", host)
	if err != nil {
		panic(err)
	}
	return &RPCClient{*cl}
}

// This is how a client invokes a "raw" RPC
// Returns true if there was an error
func (cl *RPCClient) RemoteProcedureCall(rpcid uint64, args *[]byte, reply *[]byte) bool {
	e := cl.Call("RPCHandler.Handle", &grove_common.RawRPCRequest{RpcId:rpcid, Data:*args}, reply)
	if e != nil {
		panic(e)
	}
	return false
}
