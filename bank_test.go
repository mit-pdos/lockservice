package lockservice

import (
	"fmt"
	"github.com/mit-pdos/lockservice/grove_ffi"
	"testing"
	"time"
)

func tt(t *testing.T, ck *BankClerk, amount uint64) {
	ck.SimpleTransfer(amount)
}

func ta(t *testing.T, ck *BankClerk, expected uint64) {
	x := ck.SimpleAudit()
	if x != expected {
		t.Fatalf("Audit() returned %v; expected %v", x, expected)
	}
}

func TestBank(t *testing.T) {
	fmt.Printf("Test: Bank ...\n")

	grove_ffi.SetPort(12302)
	MakeLockServer().Start()

	grove_ffi.SetPort(12303)
	MakeKVServer().Start()

	// kv_ck := MakeKVClerk("localhost:12303", nrand())
	kv_ck := MakeKVClerk(12303, nrand())
	kv_ck.Put(0, 100) // initialize bank to have 100 total

	b := MakeBank(12302, 12303)
	ck1 := MakeBankClerk(b, 0, 1, nrand())
	ck2 := MakeBankClerk(b, 0, 1, nrand())
	ck3 := MakeBankClerk(b, 0, 1, nrand())

	ta(t, ck3, 100)

	random_activity_fcn := func(ck *BankClerk) {
		for {
			ck.SimpleTransfer(nrand() % 100)
			time.Sleep(10 * time.Millisecond)
		}
	}

	go random_activity_fcn(ck1)
	go random_activity_fcn(ck2)

	time.Sleep(100 * time.Millisecond)
	ta(t, ck3, 100)
	fmt.Printf("  ... Passed\n")
}
