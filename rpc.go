package lockservice

import ()

// Returns true iff error
func CallTryLock(srv *LockServer, args *LockArgs, reply *LockReply) bool {
	go func() {
		var dummy_reply LockReply
		for {
			srv.TryLock(args, &dummy_reply)
		}
	}()

	if nondet() {
		return srv.TryLock(args, reply)
	}
	return true
}

// Returns true iff error
func CallUnlock(srv *LockServer, args *UnlockArgs, reply *UnlockReply) bool {
	go func() {
		var dummy_reply UnlockReply
		for {
			srv.Unlock(args, &dummy_reply)
		}
	}()

	if nondet() {
		return srv.Unlock(args, reply)
	}
	return true
}
