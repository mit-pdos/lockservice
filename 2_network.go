package lockservice

import ()

// Returns true iff server reported error or request "timed out"
func CallTryLock(srv *LockServer, args *TryLockRequest, reply *RPCReply) bool {
	go func() {
		dummy_reply := new(RPCReply)
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
func CallUnlock(srv *LockServer, args *UnlockRequest, reply *RPCReply) bool {
	go func() {
		dummy_reply := new(RPCReply) // Unlock and TryLock share the reply type
		for {
			srv.Unlock(args, dummy_reply)
		}
	}()

	if nondet() {
		return srv.Unlock(args, reply)
	}
	return true
}

// Returns true iff server reported error or request "timed out"
func CallPut(srv *KVServer, args *PutRequest, reply *RPCReply) bool {
	go func() {
		dummy_reply := new(RPCReply) // Unlock and TryLock share the reply type
		for {
			srv.Put(args, dummy_reply)
		}
	}()

	if nondet() {
		return srv.Put(args, reply)
	}
	return true
}

// Returns true iff server reported error or request "timed out"
func CallGet(srv *KVServer, args *GetRequest, reply *RPCReply) bool {
	go func() {
		dummy_reply := new(RPCReply) // Unlock and TryLock share the reply type
		for {
			srv.Get(args, dummy_reply)
		}
	}()

	if nondet() {
		return srv.Get(args, reply)
	}
	return true
}
