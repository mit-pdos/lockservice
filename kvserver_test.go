package lockservice

import (
	"fmt"
	"runtime"
	"testing"
	"github.com/mit-pdos/lockservice/grove_common"
	"github.com/mit-pdos/lockservice/grove_ffi"
)

func tp(t *testing.T, ck *KVClerk, k uint64, v uint64) {
	ck.Put(k, v)
}

func tg(t *testing.T, ck *KVClerk, k uint64, expected uint64) {
	x := ck.Get(k)
	if x != expected {
		t.Fatalf("get(%v) returned %v; expected %v", k, x, expected)
	}
}

func TestKVStore(t *testing.T) {
	fmt.Printf("Test: Basic seq put/get ...\n")

	runtime.GOMAXPROCS(100)

	p := MakeKVServer()

	p_handlers := make(map[uint64]grove_common.RpcFunc)
	p_handlers[KV_PUT] = p.Put
	p_handlers[KV_GET] = p.Get
	pid := grove_ffi.AllocServer(p_handlers)

	ck1 := MakeKVClerk(pid, nrand())
	tp(t, ck1, 0, 12)
	tg(t, ck1, 0, 12)
	tp(t, ck1, 0, 13)
	tg(t, ck1, 0, 13)

	tp(t, ck1, 1, 101)
	tp(t, ck1, 2, 102)
	tp(t, ck1, 3, 103)
	tp(t, ck1, 4, 104)

	tg(t, ck1, 4, 104)
	tg(t, ck1, 3, 103)
	tg(t, ck1, 2, 102)
	tg(t, ck1, 1, 101)

	fmt.Printf("  ... Passed\n")
}
