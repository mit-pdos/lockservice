package lockservice

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"runtime"
	"strconv"
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
	bhost := port("b")
	p := StartServer(phost, bhost, true)  // primary
	b := StartServer(phost, bhost, false) // backup

	ck := MakeClerk(phost, bhost, nrand())

	tl(t, ck, 0, true)
	tu(t, ck, 0, true)

	tl(t, ck, 0, true)
	tl(t, ck, 1, true)
	tu(t, ck, 0, true)
	tu(t, ck, 1, true)

	tl(t, ck, 0, true)
	tl(t, ck, 0, false)
	tu(t, ck, 0, true)
	tu(t, ck, 0, false)

	p.kill()
	b.kill()

	fmt.Printf("  ... Passed\n")
}

func TestPrimaryFail1(t *testing.T) {
	fmt.Printf("Test: Primary failure ...\n")
	runtime.GOMAXPROCS(4)

	phost := port("p")
	bhost := port("b")
	p := StartServer(phost, bhost, true)  // primary
	b := StartServer(phost, bhost, false) // backup

	ck := MakeClerk(phost, bhost, nrand())

	tl(t, ck, 0, true)

	tl(t, ck, 1, true)
	tu(t, ck, 1, true)

	tl(t, ck, 2, true)
	tl(t, ck, 2, false)

	tl(t, ck, 3, true)
	tu(t, ck, 3, true)
	tl(t, ck, 3, true)

	p.kill()

	tl(t, ck, 0, false)
	tu(t, ck, 0, true)

	tu(t, ck, 1, false)
	tl(t, ck, 1, true)

	tu(t, ck, 2, true)

	tu(t, ck, 3, true)

	b.kill()
	fmt.Printf("  ... Passed\n")
}

func TestPrimaryFail2(t *testing.T) {
	fmt.Printf("Test: Primary failure just before reply #1 ...\n")
	runtime.GOMAXPROCS(4)

	phost := port("p")
	bhost := port("b")
	p := StartServer(phost, bhost, true)  // primary
	b := StartServer(phost, bhost, false) // backup

	ck1 := MakeClerk(phost, bhost, nrand())
	ck2 := MakeClerk(phost, bhost, nrand())

	tl(t, ck1, 0, true)
	tl(t, ck1, 1, true)

	p.dying = true

	tl(t, ck2, 2, true)
	tl(t, ck1, 2, false)
	tu(t, ck2, 2, true)
	tl(t, ck1, 2, true)

	b.kill()
	fmt.Printf("  ... Passed\n")
}

func TestPrimaryFail3(t *testing.T) {
	fmt.Printf("Test: Primary failure just before reply #2 ...\n")
	runtime.GOMAXPROCS(4)

	phost := port("p")
	bhost := port("b")
	p := StartServer(phost, bhost, true)  // primary
	b := StartServer(phost, bhost, false) // backup

	ck1 := MakeClerk(phost, bhost, nrand())

	tl(t, ck1, 0, true)
	tl(t, ck1, 1, true)

	p.dying = true

	tl(t, ck1, 1, false)

	b.kill()
	fmt.Printf("  ... Passed\n")
}

func TestPrimaryFail4(t *testing.T) {
	fmt.Printf("Test: Primary failure just before reply #3 ...\n")
	runtime.GOMAXPROCS(4)

	phost := port("p")
	bhost := port("b")
	p := StartServer(phost, bhost, true)  // primary
	b := StartServer(phost, bhost, false) // backup

	ck1 := MakeClerk(phost, bhost, nrand())
	ck2 := MakeClerk(phost, bhost, nrand())

	tl(t, ck1, 0, true)
	tl(t, ck1, 1, true)

	p.dying = true

	tl(t, ck2, 1, false)

	b.kill()
	fmt.Printf("  ... Passed\n")
}

func TestPrimaryFail5(t *testing.T) {
	fmt.Printf("Test: Primary failure just before reply #4 ...\n")
	runtime.GOMAXPROCS(4)

	phost := port("p")
	bhost := port("b")
	p := StartServer(phost, bhost, true)  // primary
	b := StartServer(phost, bhost, false) // backup

	ck1 := MakeClerk(phost, bhost, nrand())
	ck2 := MakeClerk(phost, bhost, nrand())

	tl(t, ck1, 0, true)
	tl(t, ck1, 1, true)
	tu(t, ck1, 1, true)

	p.dying = true

	tu(t, ck1, 1, false)
	tl(t, ck2, 1, true)

	b.kill()
	fmt.Printf("  ... Passed\n")
}

func TestPrimaryFail6(t *testing.T) {
	fmt.Printf("Test: Primary failure just before reply #5 ...\n")
	runtime.GOMAXPROCS(4)

	phost := port("p")
	bhost := port("b")
	p := StartServer(phost, bhost, true)  // primary
	b := StartServer(phost, bhost, false) // backup

	ck1 := MakeClerk(phost, bhost, nrand())
	ck2 := MakeClerk(phost, bhost, nrand())

	tl(t, ck1, 0, true)
	tu(t, ck1, 0, true)
	tu(t, ck2, 0, false)
	tl(t, ck1, 1, true)

	p.dying = true

	tu(t, ck2, 1, true)
	tl(t, ck1, 1, true)

	b.kill()
	fmt.Printf("  ... Passed\n")
}

func TestPrimaryFail7(t *testing.T) {
	fmt.Printf("Test: Primary failure just before reply #6 ...\n")
	runtime.GOMAXPROCS(4)

	phost := port("p")
	bhost := port("b")
	p := StartServer(phost, bhost, true)  // primary
	b := StartServer(phost, bhost, false) // backup

	ck1 := MakeClerk(phost, bhost, nrand())
	ck2 := MakeClerk(phost, bhost, nrand())

	tl(t, ck1, 0, true)
	tu(t, ck1, 0, true)
	tu(t, ck2, 0, false)
	tl(t, ck1, 1, true)

	p.dying = true

	ch := make(chan bool)
	go func() {
		ok := false
		defer func() { ch <- ok }()
		tu(t, ck2, 1, true) // 2 second delay until retry
		ok = true
	}()
	time.Sleep(1 * time.Second)
	tl(t, ck1, 1, true)

	ok := <-ch
	if ok == false {
		t.Fatalf("re-sent Unlock did not return true")
	}

	tu(t, ck1, 1, true)

	b.kill()
	fmt.Printf("  ... Passed\n")
}

func TestPrimaryFail8(t *testing.T) {
	fmt.Printf("Test: Primary failure just before reply #7 ...\n")
	runtime.GOMAXPROCS(4)

	phost := port("p")
	bhost := port("b")
	p := StartServer(phost, bhost, true)  // primary
	b := StartServer(phost, bhost, false) // backup

	ck1 := MakeClerk(phost, bhost, nrand())
	ck2 := MakeClerk(phost, bhost, nrand())

	tl(t, ck1, 0, true)
	tu(t, ck1, 0, true)

	p.dying = true

	ch := make(chan bool)
	go func() {
		ok := false
		defer func() { ch <- ok }()
		tu(t, ck2, 0, false) // 2 second delay until retry
		ok = true
	}()
	time.Sleep(1 * time.Second)
	tl(t, ck1, 0, true)

	ok := <-ch
	if ok == false {
		t.Fatalf("re-sent Unlock did not return false")
	}

	tu(t, ck1, 0, true)

	b.kill()
	fmt.Printf("  ... Passed\n")
}

func TestBackupFail(t *testing.T) {
	fmt.Printf("Test: Backup failure ...\n")
	runtime.GOMAXPROCS(4)

	phost := port("p")
	bhost := port("b")
	p := StartServer(phost, bhost, true)  // primary
	b := StartServer(phost, bhost, false) // backup

	ck := MakeClerk(phost, bhost, nrand())

	tl(t, ck, 0, true)

	tl(t, ck, 1, true)
	tu(t, ck, 1, true)

	tl(t, ck, 2, true)
	tl(t, ck, 2, false)

	tl(t, ck, 3, true)
	tu(t, ck, 3, true)
	tl(t, ck, 3, true)

	b.kill()

	tl(t, ck, 0, false)
	tu(t, ck, 0, true)

	tu(t, ck, 1, false)
	tl(t, ck, 1, true)

	tu(t, ck, 2, true)

	tu(t, ck, 3, true)

	p.kill()
	fmt.Printf("  ... Passed\n")
}

func TestMany(t *testing.T) {
	fmt.Printf("Test: Multiple clients with primary failure ...\n")
	runtime.GOMAXPROCS(4)

	phost := port("p")
	bhost := port("b")
	p := StartServer(phost, bhost, true)  // primary
	b := StartServer(phost, bhost, false) // backup

	const nclients = 2
	const nlocks = 10
	done := false
	var state [nclients][nlocks]bool
	var acks [nclients]bool

	for xi := 0; xi < nclients; xi++ {
		go func(i int) {
			ck := MakeClerk(phost, bhost, nrand())
			rr := rand.New(rand.NewSource(int64(os.Getpid() + i)))
			for done == false {
				locknum := (rr.Uint64() % nlocks)
				lockname := locknum + uint64(i*1000)
				what := rr.Int() % 2
				if what == 0 {
					ck.Lock(lockname)
					state[i][locknum] = true
				} else {
					ck.Unlock(lockname)
					state[i][locknum] = false
				}
			}
			acks[i] = true
		}(xi)
	}

	time.Sleep(2 * time.Second)
	p.kill()
	time.Sleep(2 * time.Second)
	done = true
	time.Sleep(time.Second)
	ck := MakeClerk(phost, bhost, nrand())
	for xi := 0; xi < nclients; xi++ {
		if acks[xi] == false {
			t.Fatal("one client didn't complete")
		}
		for locknum := 0; locknum < nlocks; locknum++ {
			lockname := uint64(locknum + (xi * 1000))
			locked := !ck.Lock(lockname)
			if locked != state[xi][locknum] {
				t.Fatal("bad final state")
			}
		}
	}

	b.kill()
	fmt.Printf("  ... Passed\n")
}

func TestConcurrentCounts(t *testing.T) {
	fmt.Printf("Test: Multiple clients, single lock, primary failure ...\n")
	runtime.GOMAXPROCS(4)

	phost := port("p")
	bhost := port("b")
	p := StartServer(phost, bhost, true)  // primary
	b := StartServer(phost, bhost, false) // backup

	const nclients = 2
	const nlocks = 1
	done := false
	var acks [nclients]bool
	var locks [nclients][nlocks]int
	var unlocks [nclients][nlocks]int

	for xi := 0; xi < nclients; xi++ {
		go func(i int) {
			ck := MakeClerk(phost, bhost, nrand())
			rr := rand.New(rand.NewSource(int64(os.Getpid() + i)))
			for done == false {
				locknum := rr.Uint64() % nlocks
				lockname := locknum
				what := rr.Int() % 2
				if what == 0 {
					if ck.Lock(lockname) {
						locks[i][locknum]++
					}
				} else {
					if ck.Unlock(lockname) {
						unlocks[i][locknum]++
					}
				}
			}
			acks[i] = true
		}(xi)
	}

	time.Sleep(2 * time.Second)
	p.kill()
	time.Sleep(2 * time.Second)
	done = true
	time.Sleep(time.Second)
	for xi := 0; xi < nclients; xi++ {
		if acks[xi] == false {
			t.Fatal("one client didn't complete")
		}
	}
	ck := MakeClerk(phost, bhost, nrand())
	for locknum := 0; locknum < nlocks; locknum++ {
		nl := 0
		nu := 0
		for xi := 0; xi < nclients; xi++ {
			nl += locks[xi][locknum]
			nu += unlocks[xi][locknum]
		}
		locked := ck.Unlock(uint64(locknum))
		// fmt.Printf("lock=%d nl=%d nu=%d locked=%v\n",
		//   locknum, nl, nu, locked)
		if nl < nu || nl > nu+1 {
			t.Fatal("lock race 1")
		}
		if nl == nu && locked != false {
			t.Fatal("lock race 2")
		}
		if nl != nu && locked != true {
			t.Fatal("lock race 3")
		}
	}

	b.kill()
	fmt.Printf("  ... Passed\n")
}
