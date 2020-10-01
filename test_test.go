package lockservice

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	//	"math/rand"
	"os"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
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

//
// cook up a unique-ish UNIX-domain socket name
// in /var/tmp. can't use current directory since
// AFS doesn't support UNIX-domain sockets.
//
func port(suffix string) string {
	s := "/var/tmp/824-"
	s += strconv.Itoa(os.Getuid()) + "/"
	os.Mkdir(s, 0777)
	s += strconv.Itoa(os.Getpid()) + "-"
	s += suffix
	return s
}

func TestBasic(t *testing.T) {
	fmt.Printf("Test: Basic lock/unlock ...\n")

	runtime.GOMAXPROCS(4)

	phost := port("p")
	p := StartServer(phost)

	ck := MakeClerk(phost, nrand())

	tl(t, ck, 0, true)
	tu(t, ck, 0, true)

	tl(t, ck, 0, true)
	tl(t, ck, 1, true)
	tu(t, ck, 0, true)
	tu(t, ck, 1, true)

	tl(t, ck, 0, true)
	tu(t, ck, 0, true)
	tu(t, ck, 0, false)

	p.kill()

	fmt.Printf("  ... Passed\n")
}

func TestBasicConcurrent(t *testing.T) {
	fmt.Printf("Test: Basic concurrent lock/unlock ...\n")

	runtime.GOMAXPROCS(4)

	phost := port("p")
	p := StartServer(phost)

	ck1 := MakeClerk(phost, nrand())
	ck2 := MakeClerk(phost, nrand())

	val := 0

	// client 1
	incr_fcn := func(ck *Clerk, wg *sync.WaitGroup) {
		defer wg.Done()
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

	p.kill()

	fmt.Printf("  ... Passed\n")
}
