package lockservice

import (
	"./grove_ffi"
	"github.com/tchajed/marshal"
	"sync"
)

type KVServer struct {
	sv *RPCServer
	// for each lock name, is it locked?
	kvs map[uint64]uint64
}

func (ks *KVServer) put_core(args RPCVals) uint64 {
	ks.kvs[args.U64_1] = args.U64_2
	return 0
}

func (ks *KVServer) get_core(args RPCVals) uint64 {
	return ks.kvs[args.U64_1]
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
func WriteDurableKVServer(ks *KVServer) {

	// TODO: need  to make sure this doesn't overflow
	num_bytes := uint64(8 * (2*len(ks.sv.lastSeq) + 2*len(ks.sv.lastSeq) + 2*len(ks.kvs) + 3))
	e := marshal.NewEnc(num_bytes) // 4 uint64s
	EncMap(&e, ks.sv.lastSeq)
	EncMap(&e, ks.sv.lastReply)
	EncMap(&e, ks.kvs)

	grove_ffi.Write("kvdur", e.Finish())
	return
}

func ReadDurableKVServer() *KVServer {
	content := grove_ffi.Read("kvdur")
	if len(content) == 0 {
		return nil
	}
	d := marshal.NewDec(content)
	ks := new(KVServer)
	sv := new(RPCServer)
	sv.mu = new(sync.Mutex)
	sv.lastSeq = DecMap(&d)
	sv.lastReply = DecMap(&d)
	ks.kvs = DecMap(&d)
	ks.sv = sv

	return ks
}

func (ks *KVServer) Put(req *RPCRequest, reply *RPCReply) bool {
	return ks.sv.HandleRequest(
		func(args RPCVals) uint64 {
			return ks.put_core(args)
		},
		func() {
			WriteDurableKVServer(ks)
		},
		req, reply)
}

func (ks *KVServer) Get(req *RPCRequest, reply *RPCReply) bool {
	return ks.sv.HandleRequest(
		func(args RPCVals) uint64 {
			return ks.get_core(args)
		},
		func() {
			WriteDurableKVServer(ks)
		},
		req, reply)
}

func MakeKVServer() *KVServer {
	// If we alreay have some saved state, continue from there
	ks_old := ReadDurableKVServer()
	if ks_old != nil {
		return ks_old
	}

	// Otherwise, we should make a brand new object
	ks := new(KVServer)
	ks.kvs = make(map[uint64]uint64)
	ks.sv = MakeRPCServer()
	return ks
}
