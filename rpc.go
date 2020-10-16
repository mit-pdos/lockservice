package lockservice

import ()

// Returns true iff server reported error or request "timed out"
func CallTryLock(srv *LockServer, args *TryLockArgs, reply *TryLockReply) bool {
	go func() {
		dummy_reply := new(TryLockReply)
		for {
			srv.TryLock(args, dummy_reply)
		}
	}()

	if nondet() {
		return srv.TryLock(args, reply)
	}
	return true
}

// Returns true iff server reported error or request "timed out"
func CallUnlock(srv *LockServer, args *UnlockArgs, reply *UnlockReply) bool {
	go func() {
		dummy_reply := new(UnlockReply)
		for {
			srv.Unlock(args, dummy_reply)
		}
	}()

	if nondet() {
		return srv.Unlock(args, reply)
	}
	return true
}
