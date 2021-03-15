package grove_ffi

import (
	"sync"
	"github.com/mit-pdos/lockservice/grove_common"
)

// rpcHandlers describes the RpcFunc handler that will run
// when an RPC is issued to some machine (first key) to invoke
// a specific RPC number (second key).
var rpcHandlers map[uint64]map[uint64]grove_common.RpcFunc
var rpcNextHost uint64
var rpcHandlersLock sync.Mutex

func AllocServer(handlers map[uint64]grove_common.RpcFunc) uint64 {
	rpcHandlersLock.Lock()
	defer rpcHandlersLock.Unlock()

	id := rpcNextHost
	rpcNextHost = rpcNextHost + 1

	if rpcHandlers == nil {
		rpcHandlers = make(map[uint64]map[uint64]grove_common.RpcFunc)
	}

	rpcHandlers[id] = handlers
	return id
}

func GetServer(host uint64, rpc uint64) grove_common.RpcFunc {
	rpcHandlersLock.Lock()
	defer rpcHandlersLock.Unlock()

	return rpcHandlers[host][rpc]
}
