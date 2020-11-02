package lockservice

import ()

type SrvFunc func(*RPCRequest, *RPCReply) bool

// Returns true iff server reported error or request "timed out"
func CallRpc(srv SrvFunc, req *RPCRequest, reply *RPCReply) bool {
	go func() {
		dummy_reply := new(RPCReply)
		for {
			srv(req, dummy_reply)
		}
	}()

	if nondet() {
		return srv(req, reply)
	}
	return true
}
