package lockservice

import (
	"github.com/mit-pdos/goose-nfsd/lockmap"
	"sync"
)

type TxnResources struct {
	key      uint64
	oldValue uint64
}

type ParticipantServer struct {
	mu *sync.Mutex

	lockmap      *lockmap.LockMap
	kvs          map[uint64]uint64
	txns         map[uint64]TxnResources // in-progress transactions
	finishedTxns map[uint64]bool         // finished transactions
}

// Precondition: emp
// returns 0 -> Vote Yes
// returns 1 -> Vote No
func (ps *ParticipantServer) PrepareIncrease(tid, key, amount uint64) uint64 {
	ps.mu.Lock()
	_, ok := ps.txns[tid]
	if ok {
		ps.mu.Unlock()
		return 0
	}

	_, ok2 := ps.finishedTxns[tid]
	if ok2 {
		ps.mu.Unlock()
		return 1
	}

	ps.lockmap.Acquire(key)
	if amount+ps.kvs[key] < ps.kvs[key] {
		ps.lockmap.Release(key)
		ps.mu.Unlock()
		return 1 // Vote No
	}

	ps.txns[tid] = TxnResources{key: key, oldValue: ps.kvs[key]}
	// transaction now owns key
	ps.kvs[key] += amount
	// TODO(crash): save txn and state to disk
	ps.mu.Unlock()
	return 0
}

func (ps *ParticipantServer) PrepareDecrease(tid, key, amount uint64) uint64 {
	ps.mu.Lock()
	_, ok := ps.txns[tid]
	if ok {
		ps.mu.Unlock()
		return 0
	}

	_, ok2 := ps.finishedTxns[tid]
	if ok2 {
		ps.mu.Unlock()
		return 1
	}
	// NOTE: We aren't checking if we've Voted No for tid previously, so we
	// might actually end up voting yes after a bunch of retrying. That's ok,
	// because a vote no isn't a promise to never vote yes, it's just an
	// indication that the resources to become prepared weren't available.
	// Think of Vote No as the lack of a Vote Yes (manifestly, if this returns a
	// No vote, then there are no attached resources, whereas a Vote Yes will
	// include a persistent resource indicating that the kv mapsto has been put
	// in the transaction_invariant).

	ps.lockmap.Acquire(key)
	if amount > ps.kvs[key] {
		ps.lockmap.Release(key)
		ps.mu.Unlock()
		return 1 // Vote No
	}
	ps.txns[tid] = TxnResources{key: key, oldValue: ps.kvs[key]}
	ps.kvs[key] -= amount
	// TODO(crash): save txn and state to disk
	ps.mu.Unlock()
	return 0
}

func (ps *ParticipantServer) Commit(tid uint64) {
	ps.mu.Lock()
	t, ok := ps.txns[tid]
	if !ok { // invalid tid
		ps.mu.Unlock()
		return
	}
	ps.lockmap.Release(t.key)
	// TODO(crash): save txn and state to disk
	delete(ps.txns, tid)
	ps.finishedTxns[tid] = true
	ps.mu.Unlock()
}

func (ps *ParticipantServer) Abort(tid uint64) {
	ps.mu.Lock()
	t, ok := ps.txns[tid]
	if !ok { // invalid tid
		ps.mu.Unlock()
		return
	}
	ps.kvs[t.key] = t.oldValue // rollback
	ps.lockmap.Release(t.key)
	delete(ps.txns, tid)
	ps.finishedTxns[tid] = true
	// TODO(crash): save txn and state to disk
	ps.mu.Unlock()
}

func MakeParticipantServer() {
	s := new(ParticipantServer)
	s.mu = new(sync.Mutex)
	s.lockmap = lockmap.MkLockMap()
	s.kvs = make(map[uint64]uint64)
	s.txns = make(map[uint64]TxnResources)
	s.finishedTxns = make(map[uint64]bool)
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
	prepared1 := tc.s0.PrepareIncrease(tid, acc1, amount)
	prepared2 := tc.s1.PrepareDecrease(tid, acc2, amount)
	if prepared1 == 0 && prepared2 == 0 {
		// TODO(crash): Save commit on disk
		tc.s0.Commit(tid)
		tc.s1.Commit(tid)
	} else {
		tc.s0.Abort(tid)
		tc.s1.Abort(tid)
	}
}
