package lockservice

import (
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
	b := MakeBank(0, 100)
	ck1 := MakeBankClerk(b, 0, 1, nrand())
	ck2 := MakeBankClerk(b, 0, 1, nrand())
	ck3 := MakeBankClerk(b, 0, 1, nrand())

	ta(t, ck3, 100)

	random_activity_fcn := func(ck *BankClerk) {
		for {
			ck.SimpleTransfer( nrand() %100 )
			time.Sleep(10 * time.Millisecond)
		}
	}

	go random_activity_fcn(ck1)
	go random_activity_fcn(ck2)

	time.Sleep(100 * time.Millisecond)
	ta(t, ck3, 100)
}
