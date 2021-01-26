package lockservice

import (
	"sync"
)

const (
	OpDecrease = 0
	OpIncrease = 1
)

type Transaction struct {
	heldResource uint64
	oldValue uint64
	operation uint64 // 0 == decrease, 1 == increase
	amount uint64
}

type ParticipantServer struct {
	sv *RPCServer

	// TODO: lockmap
	lockmap sync.Mutex
	kvs map[uint64]uint64
	txns map[uint64]Transaction
}

// returns 0 -> Vote Yes
// returns 1 -> Vote No
func (ps *ParticipantServer) prepare_increase_core(tid uint64, key, amount uint64) uint64 {
	ps.lockmap.Lock() // Lockmap (key)
	ps.txns[tid] = Transaction{heldResource:key, oldValue:ps.kvs[key], operation:OpIncrease, amount:amount}
	// transaction now owns key
	ps.kvs[key] += amount
	// TODO(crash): save txn and state to disk
	return 0
}

func (ps *ParticipantServer) prepare_decrease_core(tid uint64, key, amount uint64) uint64 {
	// TODO(crash): check if tid is already in ps.txns; if so, then return Vote Yes
	ps.lockmap.Lock() // Lockmap (key)
	if amount > ps.kvs[key] {
		ps.lockmap.Unlock() // Lockmap (key)
	// TODO(crash): save vote on disk
		return 1 // Vote No
	}
	ps.txns[tid] = Transaction{heldResource:key, oldValue:ps.kvs[key], operation:OpDecrease, amount:amount}
	ps.kvs[key] -= amount
	// TODO(crash): save txn and state to disk
	return 0
}

func (ps *ParticipantServer) commit_core(tid uint64) {
	_, ok := ps.txns[tid]
	if !ok { // invalid tid
		return
	}
	ps.lockmap.Unlock() // t.heldResource
	delete(ps.txns, tid)
}

func (ps *ParticipantServer) abort_core(tid uint64) {
	t, ok := ps.txns[tid]
	if !ok { // invalid tid
		return
	}
	ps.kvs[t.heldResource] = t.oldValue // rollback
	ps.lockmap.Unlock() // t.heldResource
	delete(ps.txns, tid)
}

type TransactionCoordinator struct {
	s0 *ParticipantServer
	s1 *ParticipantServer
}

// transfers between acc1 on s0 and acc2 on s1
// could also shard key-space
func (tc *TransactionCoordinator) doTransfer(tid, acc1, acc2, amount uint64) {
	// FIXME: these should go over rpc instead of directly call the core funcs.
	// Requires making RPCVals have space for another uint64, and fixing up
	// the proof.
	prepared1 := tc.s0.prepare_increase_core(tid, acc1, amount)
	prepared2 := tc.s0.prepare_decrease_core(tid, acc1, amount)
	if prepared1 == 0 && prepared2 == 0 {
		// TODO(crash): Save commit on disk
		tc.s0.commit_core(tid)
		tc.s1.commit_core(tid)
	} else {
		tc.s0.abort_core(tid)
		tc.s1.abort_core(tid)
	}
}
