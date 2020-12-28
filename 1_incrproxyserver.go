package lockservice

import (
// ffi "./grove_ffi"
// "github.com/tchajed/marshal"
// "fmt"
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

func (is *IncrProxyServer) proxy_increment_core(seq uint64, args RPCVals) uint64 {
	// TODO: unlike the incr server that uses kv backend, making this crash-safe
	// will *require* saving the seqno used for the request before sending it out.
	return 0
}
