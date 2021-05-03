package grove_ffi

import (
	"fmt"
	"github.com/mit-pdos/lockservice/grove_common"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

var port uint64
var config map[uint64]string

type HostName = uint64

// this is NOT exposed in the FFI. This just allows us to have StartServer()
// serve at different ports on the same host without exposing port numbers to
// proof-land
func SetPort(p uint64) {
	port = p
}

func SetConfig(c map[uint64]string) {
	config = c
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
	s.rpcHandlers[req.RpcId](req.Data, rep)
	return nil
}

// starts running an RPC server with the given functions at the corresponding
// endpoints; returns immediately, starting server in background
func StartRPCServer(handlers map[uint64]grove_common.RawRpcFunc) {
	s := &RPCHandler{}

	s.rpcHandlers = handlers

	serv := rpc.NewServer()
	serv.Register(s)

	// XXX: https://github.com/golang/go/issues/13395
	// ===== workaround ==========
	oldMux := http.DefaultServeMux
	mux := http.NewServeMux()
	http.DefaultServeMux = mux
	// ===========================

	serv.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	// ===== workaround ==========
	http.DefaultServeMux = oldMux
	// ===========================

	l, e := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if e != nil {
		panic(e)
	}
	go func() {
		log.Fatal(http.Serve(l, mux))
	}()
}

type RPCClient struct {
	cl *rpc.Client
}

func MakeRPCClient(host uint64) *RPCClient {
	cl, err := rpc.DialHTTP("tcp", config[host])
	if err != nil {
		panic(err)
	}
	return &RPCClient{cl}
}

// This is how a client invokes a "raw" RPC
// Returns true if there was an error
func (cl *RPCClient) Call(rpcid uint64, args []byte, reply *[]byte) bool {
	*reply = make([]byte, 0)
	e := cl.cl.Call("RPCHandler.Handle", &grove_common.RawRPCRequest{RpcId: rpcid, Data: args}, reply)
	if e != nil {
		panic(e)
	}
	return false
}
