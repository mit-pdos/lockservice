package lockservice

import (
	"fmt"
	"github.com/mit-pdos/lockservice/grove_ffi"
	"runtime"
	"testing"
	"log"
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
	log.Printf("Starting kvserver\n")
	// need to be careful not to reuse ports across tests
	grove_ffi.SetPort(12301)
	go p.Start()


	ck1 := MakeKVClerk("localhost:12301", nrand())
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
