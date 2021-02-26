package lockservice

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"runtime"
	"testing"
	"time"
	"github.com/mit-pdos/lockservice/grove_common"
	"github.com/mit-pdos/lockservice/grove_ffi"
)

func nrand() uint64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := crand.Int(crand.Reader, max)
	x := bigx.Uint64()
	return x
}

func tl(t *testing.T, ck *Clerk, lockname uint64, expected bool) {
	x := ck.Lock(lockname)
	if x != expected {
		t.Fatalf("Lock(%v) returned %v; expected %v", lockname, x, expected)
	}
}

func tu(t *testing.T, ck *Clerk, lockname uint64, expected bool) {
	x := ck.Unlock(lockname)
	if x != expected {
		t.Fatalf("Unlock(%v) returned %v; expected %v", lockname, x, expected)
	}
}

func TestBasicConcurrent(t *testing.T) {
	fmt.Printf("Test: Basic concurrent lock/unlock ...\n")

	runtime.GOMAXPROCS(100)

	p := MakeLockServer()

	p_handlers := make(map[uint64]grove_common.RpcFunc)
	p_handlers[LOCK_TRYLOCK] = p.TryLock
	p_handlers[LOCK_UNLOCK] = p.Unlock
	pid := grove_ffi.AllocServer(p_handlers)

	ck1 := MakeClerk(pid, nrand())
	ck2 := MakeClerk(pid, nrand())

	val := 0

	// client 1
	incr_fcn := func(ck *Clerk, wg *sync.WaitGroup) {
		defer wg.Done()
		fmt.Println("trying to get lock")
		tl(t, ck, 0, true)
		fmt.Println("got lock")
		tmp := val
		tmp = tmp + 1
		time.Sleep(00 * time.Millisecond)
		val = tmp
		tu(t, ck, 0, true)
		fmt.Println("unlocked")
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go incr_fcn(ck1, &wg)
	go incr_fcn(ck2, &wg)

	wg.Wait()

	if val != 2 {
		t.Fatalf("val is %v; expected %v", val, 2)
	}

	fmt.Printf("  ... Passed\n")
}
