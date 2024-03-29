package lockservice

import (
	"github.com/mit-pdos/lockservice/grove_common"
	"github.com/mit-pdos/lockservice/grove_ffi"
	"github.com/tchajed/marshal"
	"sync"
)

// Basic durable KV server that saves the entire state on disk at the end of
// every request. This should probably be made somewhat obsolete by gokv.

type DurableKVServer struct {
	mu  *sync.Mutex
	sv  *RPCServer
	kvs map[uint64]uint64
}

func (ks *DurableKVServer) put_core(args grove_common.RPCVals) uint64 {
	ks.kvs[args.U64_1] = args.U64_2
	return 0
}

func (ks *DurableKVServer) get_core(args grove_common.RPCVals) uint64 {
	return ks.kvs[args.U64_1]
}

func (ks *DurableKVServer) cas_core(args grove_common.RPCVals) uint64 {
	r := ks.kvs[args.U64_1]
	if r == args.U64_2 {
		ks.kvs[args.U64_1] = args.U64_2 // FIXME: U64_3
		return 1
	}
	return 0
}

// requires (2n + 1) uint64s worth of space in the encoder
func EncMap(e *marshal.Enc, m map[uint64]uint64) {
	e.PutInt(uint64(len(m)))
	for key, value := range m {
		e.PutInt(key)
		e.PutInt(value)
	}
}

func DecMap(d *marshal.Dec) map[uint64]uint64 {
	sz := d.GetInt()
	m := make(map[uint64]uint64)
	var i = uint64(0)
	for i < sz {
		k := d.GetInt()
		v := d.GetInt()
		m[k] = v
		i = i + 1
	}
	return m
}

// For now, there is only one kv server in the whole world
// Assume it's in file "kvdur"
func WriteDurableKVServer(ks *DurableKVServer) {

	// TODO: need  to make sure this doesn't overflow
	num_bytes := uint64(8 * (2*len(ks.sv.lastSeq) + 2*len(ks.sv.lastSeq) + 2*len(ks.kvs) + 3))
	e := marshal.NewEnc(num_bytes) // 4 uint64s
	EncMap(&e, ks.sv.lastSeq)
	EncMap(&e, ks.sv.lastReply)
	EncMap(&e, ks.kvs)

	grove_ffi.Write("kvdur", e.Finish())
	return
}

func ReadDurableKVServer() *DurableKVServer {
	content := grove_ffi.Read("kvdur")
	if len(content) == 0 {
		return nil
	}
	d := marshal.NewDec(content)
	ks := new(DurableKVServer)
	sv := new(RPCServer)
	sv.lastSeq = DecMap(&d)
	sv.lastReply = DecMap(&d)
	ks.kvs = DecMap(&d)
	ks.sv = sv
	ks.mu = new(sync.Mutex)

	return ks
}

func (ks *DurableKVServer) Put(req *grove_common.RPCRequest, reply *grove_common.RPCReply) {
	ks.mu.Lock()
	ks.sv.HandleRequest(ks.put_core, req, reply)
	WriteDurableKVServer(ks)
	ks.mu.Unlock()
	return
}

func (ks *DurableKVServer) Get(req *grove_common.RPCRequest, reply *grove_common.RPCReply) {
	ks.mu.Lock()
	ks.sv.HandleRequest(ks.get_core, req, reply)
	WriteDurableKVServer(ks) // for updating reply table
	ks.mu.Unlock()
	return
}

func (ks *DurableKVServer) CAS(req *grove_common.RPCRequest, reply *grove_common.RPCReply) {
	ks.mu.Lock()
	ks.sv.HandleRequest(ks.cas_core, req, reply)
	WriteDurableKVServer(ks) // for updating reply table
	ks.mu.Unlock()
	return
}

func MakeDurableKVServer() *DurableKVServer {
	// If we alreay have some saved state, continue from there
	ks_old := ReadDurableKVServer()
	if ks_old != nil {
		return ks_old
	}

	// Otherwise, we should make a brand new object
	ks := new(DurableKVServer)
	ks.mu = new(sync.Mutex)
	ks.kvs = make(map[uint64]uint64)
	ks.sv = MakeRPCServer()
	return ks
}

// Does not return
func (ks *DurableKVServer) Start() {
	handlers := make(map[uint64]grove_common.RawRpcFunc)
	// handlers[KV_PUT] = ks.Put
	// handlers[KV_GET] = ks.Get
	grove_ffi.StartRPCServer(handlers)
}
